/*
房间开赛前
*/

package engine

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	. "kelei.com/utils/common"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

var (
	goodCard_KingCount = 5
	goodCard_TwoCount  = 8
)

/*
推送房间的状态信息
push:Matching_Push,64$1$0$5$0|||||#等待席#当前轮次
*/
func (r *Room) matchingPush(user *User) {
	//获取此房间的信息
	userids, statuss := r.getRoomMatchingInfo()
	if user != nil {
		userids = []string{*user.getUserID()}
	}
	pushMessageToUsers("Matching_Push", statuss, userids)
	r.pushJudgment("Matching_Push", statuss[0])
}

/*
推送海选赛匹配信息
push:MatchingHXS_Push,64$1$0$5$0|||||#等待席#当前轮次
*/
func (r *Room) matchingHXSPush(user *User) {
	//获取此房间的信息
	userids, statuss := r.getRoomMatchingInfo()
	if user != nil {
		userids = []string{*user.getUserID()}
	}
	pushMessageToUsers("MatchingHXS_Push", statuss, userids)
}

/*
所有人开启准备倒计时
*/
func (r *Room) setOutCountDown() {
	matchid := r.GetMatchID()
	if matchid == Match_HYTW || matchid == Match_HXS {
		return
	}
	users := r.getUsers()
	for _, user := range users {
		user.countDown_setOut(time.Second * 20)
	}
}

//初始化房间中的玩家信息
func (r *Room) initUsersInfo() {
	for i, user := range r.users {
		user.setStatus(UserStatus_NoPass)
		user.setIndex(i)
		user.close_countDown_playCard()
	}
}

//开局
func (r *Room) opening() {
	//是录制版
	if r.getGameRule() == GameRule_Record {
		return
	}
	//海选赛延迟开赛
	if r.GetMatchID() == Match_HXS {
		go func() {
			defer func() {
				if p := recover(); p != nil {
					errInfo := fmt.Sprintf("opening : { %v }", p)
					logger.Errorf(errInfo)
				}
			}()
			time.Sleep(time.Second * 4)
			//比赛开局
			r.match_Opening()
		}()
	} else {
		//比赛开局
		r.match_Opening()
	}
}

//发牌
func (r *Room) deal() {
	r.match_Opening()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				errInfo := fmt.Sprintf("deal : { %v }", p)
				logger.Errorf(errInfo)
			}
		}()
		pushMessageToUsers("Deal_Push", []string{"1"}, r.getUserIDs())
	}()
}

//是否人齐
func (r *Room) userEnough() bool {
	return r.getUserCount() >= pcount
}

//开局处理
func (r *Room) opening_handle() {
	//检测开赛
	r.check_start()
}

//开启加倍
func (r *Room) openDouble() {
	go func() {
		if r.getGameType() == GAMETYPE_DOUBLE {
			time.Sleep(time.Second * 5)
		}
		r.doubleDefaultHandle()
		//开启明牌
		r.openOpenHand()
	}()
}

//加倍的默认处理
func (r *Room) doubleDefaultHandle() {
	users := r.getUsers()
	for _, user := range users {
		if user.getDoubleStatus() == HANDLE_NOP {
			if r.getGameType() == GAMETYPE_DOUBLE {
				NoDouble(user.getArgs())
			} else {
				user.setDoubleStatus(HANDLE_NO)
			}
		}
	}
}

//开启明牌
func (r *Room) openOpenHand() {
	go func() {
		time.Sleep(time.Second * 5)
		r.openHandDefaultHandle()
	}()
}

//名牌的默认处理
func (r *Room) openHandDefaultHandle() {
	users := r.getUsers()
	for _, user := range users {
		if user.getOpenHandStatus() == HANDLE_NOP {
			NoOpenHand(user.getArgs())
		}
	}
	time.Sleep(time.Second * 1)
	r.checkStart()
}

//检测开牌
func (r *Room) checkStart() {
	//没有地主，不能开牌
	if r.getLandlord() == nil {
		return
	}
	r.opening_handle()
	r.setMatchingStatus(MatchingStatus_Run)
	pushMessageToUsers("Begin_Push", []string{"1"}, r.getUserIDs())
	r.start()
}

//设置叫地主的玩家
func (r *Room) setCallLandlordUser() {
	users := r.getUsers()
	var canCallLandlordUser *User
	if r.getCanCallLandlordUser() == nil {
		rd := Random(0, pcount)
		logger.Debugf("第一场叫地主:%d", rd)
		canCallLandlordUser = users[rd]
	} else {
		index := r.getCanCallLandlordUser().getIndex()
		index += 1
		if index >= pcount {
			index = 0
		}
		logger.Debugf("轮流叫地主:%d", index)
		canCallLandlordUser = users[index]
	}
	r.setCanCallLandlordUser(canCallLandlordUser)
	r.setCanHandleUserAndSetCountDown(canCallLandlordUser, HANDLETYPE_CALL)
}

//开启叫地主
func (r *Room) openCallLandlord() {
	r.setCallLandlordUser()
}

//开启玩家操作(明、叫、抢、加倍)
func (r *Room) openUsersHandle() {
	//设置房间处理玩家操作的状态
	r.SetRoomStatus(RoomStatus_Handle)
	//开启叫地主
	r.openCallLandlord()
}

//展示本局任务
func (r *Room) showCouncilTask() {
	task := taskSystem.randomGetTask()
	r.setCouncilTask(task)
}

//比赛开局(发牌中)
func (r *Room) match_Opening() {
	if !r.userEnough() {
		return
	}
	if r.GetRoomStatus() == RoomStatus_Deal {
		return
	}
	//展示本局任务
	r.showCouncilTask()
	//设置房间为发牌状态
	r.SetRoomStatus(RoomStatus_Deal)
	//将玩家比赛的信息同步到数据库
	r.insertUsersInfo()
	//开赛更新玩家数据
	r.updateUserInfo()
	//初始化房间中的玩家信息
	r.initUsersInfo()
	//推送开赛牌局
	r.Opening_Push()
	//	go func() {
	//		defer func() {
	//			if p := recover(); p != nil {
	//				errInfo := fmt.Sprintf("match_Opening : { %v }", p)
	//				logger.Errorf(errInfo)
	//			}
	//		}()
	//		//等待发牌
	//		if r.GetMatchID() == Match_JD {
	//			time.Sleep(time.Second * 1)
	//		} else if r.GetMatchID() == Match_HYTW {
	//			time.Sleep(time.Second * 1)
	//		} else {
	//			time.Sleep(time.Second * 1)
	//		}
	//		//开启玩家操作(明、叫、抢、加倍)
	//		r.openUsersHandle()
	//	}()
}

//将玩家比赛的信息同步到数据库
func (r *Room) insertUsersInfo() {
	users := r.getUsers()
	for _, user := range users {
		user.insertUserInfo()
	}
}

//推送开赛牌局
func (r *Room) Opening_Push() {
	//生成所有人的牌
	cardsList := []CardList{}
	//	count := 0
	//	CostTime(func() {
	//		for {
	//			count++
	//			cardsList, _ = r.generateCards()
	//			if len(cardsList) > 0 {
	//				break
	//			}
	//		}
	//	}, 100, "发牌")
	//	fmt.Println("100场发牌总次数：", count)
	//	return
	for {
		cardsList = r.generateCards()
		if len(cardsList) > 0 {
			break
		}
	}
	users := r.getUsers()
	userids := []string{}
	messages := []string{}
	for i, user := range users {
		//设置玩家牌面
		user.setCards(cardsList[i])
		userids = append(userids, *user.userid)
		buffer := bytes.Buffer{}
		for _, card := range cardsList[i] {
			buffer.WriteString(strconv.Itoa(card.ID))
			buffer.WriteString("|")
		}
		str := buffer.String()
		str = str[0 : len(str)-1]
		messages = append(messages, str)
	}
	//给所有人推送牌局
	pushMessageToUsers("Opening_Push", messages, userids)
	time.Sleep(time.Millisecond * 10)
	//给所有人推牌的数量
	for _, user := range r.getUsers() {
		user.pushSurplusCardCount()
	}
	/*
		Opening_Judgment_Push(给裁判推送所有人的牌局)
		push:0|0|3...$
	*/
	r.pushJudgment("Opening_Judgment_Push", strings.Join(messages, "$"))
}

//检测开赛
func (r *Room) check_start() {
	//是否可以开赛
	canStart := true
	//是否可以开牌了的逻辑写在这
	if canStart {
		//不是录制版
		if r.getGameRule() != GameRule_Record {
			//开赛
			r.start()
		}
	}
}

//开始
func (r *Room) begin() {
	if frame.GetMode() == frame.MODE_RELEASE {
		defer func() {
			if p := recover(); p != nil {
				logger.Errorf("[recovery] begin : %v", p)
			}
		}()
	}
	//开启玩家操作(明、叫、抢、加倍)
	r.openUsersHandle()
}

//开赛
func (r *Room) start() {
	if !r.userEnough() {
		return
	}
	//设置比赛开始
	r.SetRoomStatus(RoomStatus_Match)
	//设置第一个出牌人
	r.setFirstController()
}

//检测有一个人网络掉线了,比赛暂停
func (r *Room) checkMatchPause(user *User) {
	//裁判端
	if r.getGameRule() == GameRule_Record {
		r.pause()
		/*
			push:[Online_Push] userid|status
			des:status=0离线 status=1回来
		*/
		user.setOnline(false)
		r.pushJudgment("Online_Push", fmt.Sprintf("%s|%d", *user.getUserID(), 0))
	}
}

//暂停
func (r *Room) pause() {
	if !r.isMatching() {
		return
	}
	pushMessageToUsers("Pause_Push", []string{"1"}, r.getUserIDs())
	r.pushJudgment("Pause_Push", "1")
	r.setMatchingStatus(MatchingStatus_Pause)
	logger.Debugf("暂停")
	for _, user := range r.getUsers() {
		if user != nil {
			user.pause()
		}
	}
}

//恢复
func (r *Room) resume() {
	pushMessageToUsers("Resume_Push", []string{"1"}, r.getUserIDs())
	r.pushJudgment("Resume_Push", "1")
	r.setMatchingStatus(MatchingStatus_Run)
	logger.Debugf("恢复")
	for _, user := range r.getUsers() {
		if user != nil {
			user.resume()
		}
	}
}

//解散牌局
func (r *Room) dissolve() {
	r.reset()
	users := r.getUsers()
	for _, user := range users {
		user.reset()
		user.setStatus(UserStatus_Setout)
		user.push("Dissolve_Push", &Res_Succeed)
	}
	r.deleteUsersInfo()
	r.SetRoomStatus(RoomStatus_Setout)
	r.setMatchingStatus(MatchingStatus_Over)
	logger.Debugf("解散牌局")
}

//获取王的数量
func (r *Room) getCountWithKing(cards CardList) int {
	count := 0
	for _, card := range cards {
		if card.Priority == Priority_SKing || card.Priority == Priority_BKing {
			count++
		} else {
			break
		}
	}
	return count
}

//获取2的数量
func (r *Room) getCountWithTwo(cards CardList) int {
	count := 0
	for _, card := range cards {
		if card.Priority >= Priority_Two {
			if card.Priority == Priority_Two {
				count++
			}
		} else {
			break
		}
	}
	return count
}

//获取大于10的牌的数量
func (r *Room) getCountWithGreaterThanTen(cards CardList) int {
	count := 0
	for _, card := range cards {
		if card.Priority > Priority_Ten {
			count++
		} else {
			break
		}
	}
	return count
}

//
func (r *Room) GenerateCards() {
	r.generateCards()
}

//不洗牌生成牌
func (r *Room) generateCardsWithNoWash() []CardList {
	rr := rand.New(rand.NewSource(time.Now().UnixNano()))
	zhadanCount := Random(2, 5)
	prioritys := rr.Perm(Priority_Two)
	zhadanPrioritys := prioritys[:zhadanCount]
	fmt.Println(zhadanPrioritys, zhadanCount, zhadanCount*4)
	zhadanMap := map[int][]Card{}
	for _, priority := range zhadanPrioritys {
		zhadanMap[priority+1] = []Card{}
	}
	surplusCardPool := []Card{}
	for _, card := range cardPool {
		if IndexIntOf(zhadanPrioritys, card.Priority-1) < 0 {
			surplusCardPool = append(surplusCardPool, card)
		} else {
			zhadanMap[card.Priority] = append(zhadanMap[card.Priority], card)
		}
	}
	cardLists := make([]CardList, pcount)
	for i := 0; i < pcount; i++ {
		cardLists[i] = CardList{}
	}
	for _, zhadanPriority := range zhadanPrioritys {
		rdUserIndex := Random(0, 3)
		cardLists[rdUserIndex] = append(cardLists[rdUserIndex], zhadanMap[zhadanPriority+1]...)
	}
	var index = 0
	indexs := rr.Perm(len(surplusCardPool))
	baseCards := CardList{}
	for i := 0; i < len(indexs); i++ {
		card := surplusCardPool[indexs[i]]
		if i < len(indexs)-3 {
			if len(cardLists[0]) < perCapitaCardCount {
				cardLists[0] = append(cardLists[0], card)
			} else if len(cardLists[1]) < perCapitaCardCount {
				cardLists[1] = append(cardLists[1], card)
			} else if len(cardLists[2]) < perCapitaCardCount {
				cardLists[2] = append(cardLists[2], card)
			}
		} else {
			baseCards = append(baseCards, card)
		}
		index += 1
	}
	//生成底牌
	sort.Sort(sort.Reverse(baseCards))
	r.setBaseCards(baseCards)
	return cardLists
}

//随机生成牌
func (r *Room) generateCardsWithRandom() []CardList {
	rr := rand.New(rand.NewSource(time.Now().UnixNano()))
	indexs := []int{}
	//初始化所有玩家默认的牌列表
	cardLists := make([]CardList, pcount)
	var index = 0
	for i := 0; i < pcount; i++ {
		cardLists[i] = CardList(make(CardList, perCapitaCardCount))
	}
	indexs = rr.Perm(cardPoolSize)
	for j := 0; j < perCapitaCardCount; j++ {
		for i := 0; i < pcount; i++ {
			card := cardPool[indexs[index]]
			cardLists[i][j] = card
			index = index + 1
		}
	}
	//生成底牌
	cards := CardList{}
	for i := 0; i < 3; i++ {
		cards = append(cards, cardPool[indexs[index]])
		index = index + 1
	}
	sort.Sort(sort.Reverse(cards))
	r.setBaseCards(cards)
	return cardLists
}

//生成所有人的牌
func (r *Room) generateCards() []CardList {
	//初始化所有玩家默认的牌列表
	cardLists := make([]CardList, pcount)
	if GetCardMode() >= 0 {
		r.SetCardMode(GetCardMode())
	}
	//生成玩家的牌
	if r.GetCardMode() == CARDMODE_RANDOM {
		cardLists = r.generateCardsWithRandom()
	} else if r.GetCardMode() == CARDMODE_NOWASH {
		cardLists = r.generateCardsWithNoWash()
	}
	//排序玩家的牌
	for i := 0; i < pcount; i++ {
		sort.Sort(cardLists[i])
		for k := 0; k < len(cardLists[i]); k++ {
			cardLists[i][k].Index = k
		}
	}
	return cardLists
}

//处理生成的牌
func (r *Room) handleGenerateCards(cardLists []CardList) []CardList {
	userStrs := make([]string, pcount)
	//	for i := 0; i < pcount; i++ {
	//		if i%2 == 0 {
	//			userStrs[i] = "2-1-1-12|2-1-1-12|2-1-1-12"
	//		} else {
	//			userStrs[i] = "10-9-1-7|11-10-1-8|12-11-1-9|13-12-1-10|14-13-1-11"
	//		}
	//	}
	for i := 0; i < pcount; i++ {
		if userStrs[i] == "" {
			cardLists[i] = cardLists[i][:cardCount]
			continue
		}
		arr := strings.Split(userStrs[i], "|")
		cards := CardList{}
		for i, cardInfo := range arr {
			cardInfo = cardInfo
			arr2 := strings.Split(cardInfo, "-")
			id_, value, suit, priority := arr2[0], arr2[1], arr2[2], arr2[3]
			id, _ := strconv.Atoi(id_)
			v, _ := strconv.Atoi(value)
			s, _ := strconv.Atoi(suit)
			p, _ := strconv.Atoi(priority)
			cards = append(cards, Card{id, v, s, p, i})
		}
		cardLists[i] = cards
	}
	return cardLists
}

//获取第一个出牌的人
func (r *Room) getFirstController() (user *User) {
	return r.firstController
}

//设置第一个出牌人
func (r *Room) setFirstController() {
	//获取第一个出牌的人
	user := r.getLandlord()
	//设置第一个出牌人
	r.setController(user, SetController_NewCycle)
}

//开赛更新玩家数据
func (r *Room) updateUserInfo() {
	if r.GetMatchID() == Match_JD {
		roomData, err := redis.Ints(r.GetRoomData("expendIngot", "integral", "charm"))
		logger.CheckFatal(err, "updateUserInfo:1")
		expendIngot, integral, charm := roomData[0], roomData[1], roomData[2]
		users := r.users
		for _, user := range users {
			user.updateIngot(-expendIngot, 3)
			user.beginGetItemEffect()
			user.beginUpdateUserInfo(map[string]int{"integral": integral, "charm": charm})
		}
	}
}
