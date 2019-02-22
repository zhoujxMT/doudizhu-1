/*
房间
*/

package engine

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

/*
游戏规则
默认版{
	1. 出牌时间15秒
	2. 自动出牌1次托管
}
录制版{
	1. 出牌时间30秒
	2. 自动出牌不托管
}
*/
const (
	GameRule_Normal = iota //默认版
	GameRule_Record        //录制版
)

const (
	Match_JD   = iota //经典
	Match_HYTW        //好友同玩
	Match_HXS         //海选赛
)

const (
	CARDMODE_RANDOM = iota //随机
	CARDMODE_NOWASH        //不洗牌
)

const (
	GAMETYPE_REGULAR = iota //常规赛
	GAMETYPE_DOUBLE         //加倍赛
)

const (
	HANDLETYPE_CALL = iota //叫地主
	HANDLETYPE_RUSH        //抢地主
)

const (
	RoomType_Primary      = iota //初级
	RoomType_Intermediate        //中级
	RoomType_Advanced            //高级
	RoomType_Master              //大师
	RoomType_Tribute             //进贡
)

const (
	SetController_NewCycle = iota //新一轮
	SetController_Press           //压牌
	SetController_Pass            //要不了
	SetController_NoChange        //没有变化
	SetController_Liuju           //流局
)

const (
	RoomStatus_Setout = iota //准备
	RoomStatus_Deal          //发牌(可明牌)
	RoomStatus_Handle        //叫地主、抢地主、加倍(可明牌)
	RoomStatus_Liuju         //流局
	RoomStatus_Match         //开赛
)

const (
	MatchingStatus_Run   = iota //进行中
	MatchingStatus_Pause        //暂停
	MatchingStatus_Over         //结束
)

const (
	PlayWaitTime      = 10 //要不起的等待时间
	PlayWaitTime_Long = 20 //其它的等待时间
)

type Room struct {
	id                    string            //id
	matchid               int               //比赛类型
	roomtype              int               //房间类型
	pcount                int               //人数
	status                int               //房间状态
	matchingStatus        int               //开赛后的状态
	users                 []*User           //玩家列表
	userids               []string          //玩家UserID集合
	idleusers             map[string]*User  //未落座玩家列表
	idleuserids           []string          //未落座玩家UserID集合
	cuser                 *User             //牌权的玩家
	cards                 []Card            //当前牌
	cardsuser             *User             //当前牌的玩家
	playTime              int               //出牌的次数
	playRound             int               //出牌的轮次
	users_cards           map[string]string //当前轮所有人的出牌信息
	inning                int               //当前局数
	innings               int               //总局数
	inningRegular         int               //常规赛局数
	setCtlMsg             []string          //设置牌权的内容,推送残局的时候用
	surplusBKingCount     int               //剩余大王数量
	surplusSKingCount     int               //剩余小王数量
	surplusTwoCount       int               //剩余2数量
	cardinality           int               //基数
	baseScore             int               //底分
	multiple              int               //倍数
	liujuMultiple         int               //流局倍数
	playWaitTime          int               //要不起等待时间
	playWaitTime_Long     int               //其它等待时间
	gameRule              int               //游戏规则
	firstController       *User             //第一个出牌的人
	judgmentUser          *User             //裁判
	records               []*string         //所有的记录（回放用）
	dealMode              int               //发牌模式
	cardMode              int               //牌的模式(随机、不洗牌)
	gameType              int               //游戏类型
	baseCards             []Card            //底牌
	landlord              *User             //地主
	farmers               []*User           //农民
	canHandleUser         *User             //当前可操作的玩家
	canCallLandlordUser   *User             //可叫地主的玩家
	landlordPlayCardCount int               //地主出牌次数
	farmerPlayCardCount   int               //农民出牌次数
	councilTask           *Task             //本局任务
	usersVideoIntegral    []int             //玩家积分列表
	springStatus          int               //春天的状态(0无1春天2反春)
}

func (r *Room) GetRoomID() *string {
	return &r.id
}

func (r *Room) SetRoomID(roomid string) {
	r.id = roomid
}

//根据玩法规则配置房间
func (r *Room) configRoomByGameRule() {
	r.playWaitTime = PlayWaitTime
	r.playWaitTime_Long = PlayWaitTime_Long
	r.setGameRule(r.GetGameRuleConfig())
	if r.getGameRule() == GameRule_Record {
		r.playWaitTime = 10
		r.playWaitTime_Long = 20
	}
}

//重置
func (r *Room) reset() {
	r.userids = nil
	r.setPlayTime(0)
	r.setPlayRound(0)
	r.setSurplusBKingCount(4)
	r.setSurplusSKingCount(4)
	r.setSurplusTwoCount(16)
	r.setControllerUser(nil)
	r.setCurrentCards([]Card{})
	r.setCurrentCardsUser(nil)
	r.setSetCtlMsg([]string{})
	r.setBaseScore(0)
	r.setMultiple(1)
	r.setLandlord(nil)
	r.setLandlordPlayCardCount(0)
	r.setFarmerPlayCardCount(0)
	for _, user := range r.getUsers() {
		if user != nil {
			user.resume()
		}
	}
	r.users_cards = make(map[string]string, pcount)
}

//设置房间的基础信息
func (r *Room) setRoomBaseInfo() {
	allRoomData := *r.getAllRoomData()
	arrAllRoomData := strings.Split(allRoomData, "|")
	for _, roomData := range arrAllRoomData {
		arrRoomData_s := strings.Split(roomData, "$")
		arrRoomData := StrArrToIntArr(arrRoomData_s)
		roomType, _, multiple := arrRoomData[0], arrRoomData[1], arrRoomData[2]
		if roomType == r.GetRoomType() {
			r.setMultiple(multiple)
			break
		}
	}
}

//是否赛前玩家操作中
func (r *Room) isHandling() bool {
	if r.GetRoomStatus() == RoomStatus_Handle {
		return true
	}
	return false
}

//是否正在比赛
func (r *Room) isMatching() bool {
	if r.GetRoomStatus() == RoomStatus_Setout {
		return false
	}
	return true
}

//获取游戏规则
func (r *Room) getGameRule() int {
	return r.gameRule
}

//设置游戏规则
func (r *Room) setGameRule(gameRule int) {
	r.gameRule = gameRule
}

//获取发牌模式
func (r *Room) getDealMode() int {
	return r.dealMode
}

//设置发牌模式
func (r *Room) setDealMode(dealMode int) {
	r.dealMode = dealMode
}

//获取牌的模式
func (r *Room) GetCardMode() int {
	return r.cardMode
}

//设置牌的模式
func (r *Room) SetCardMode(cardMode int) {
	r.cardMode = cardMode
}

//获取游戏模式
func (r *Room) getGameType() int {
	return r.gameType
}

//设置游戏模式
func (r *Room) setGameType(gameType int) {
	r.gameType = gameType
}

//获取底牌
func (r *Room) getBaseCards() []Card {
	return r.baseCards
}

//设置底牌
func (r *Room) setBaseCards(baseCards []Card) {
	r.baseCards = baseCards
}

//获取地主
func (r *Room) getLandlord() *User {
	return r.landlord
}

//设置地主
func (r *Room) setLandlord(landlord *User) {
	r.landlord = landlord
}

//获取农民
func (r *Room) getFarmers() []*User {
	return r.farmers
}

//设置农民
func (r *Room) setFarmers(users []*User) {
	r.farmers = users
}

//获取当前可操作的玩家
func (r *Room) getCanHandleUser() *User {
	return r.canHandleUser
}

/*
设置当前可操作的玩家
push:Handle_Push,userid,操作类型,当前底分,赛制
des:操作类型(0叫地主 1抢地主)
	赛制(0常规赛 1加倍赛)
*/
func (r *Room) setCanHandleUser(canHandleUser *User, handleType int) {
	r.canHandleUser = canHandleUser
	message := fmt.Sprintf("%s,%d,%d,%d", *canHandleUser.getUserID(), handleType, r.getBaseScore(), r.getGameType())
	pushMessageToUsers("Handle_Push", []string{message}, r.getUserIDs())
	r.pushJudgment("Handle_Push", message)
}

/*
设置当前可操作的玩家并设置倒计时
*/
func (r *Room) setCanHandleUserAndSetCountDown(canHandleUser *User, handleType int) {
	canHandleUser.countDown_handle(time.Second * 10)
	r.setCanHandleUser(canHandleUser, handleType)
}

//获取可以叫地主的玩家
func (r *Room) getCanCallLandlordUser() *User {
	return r.canCallLandlordUser
}

//设置可以叫地主的玩家
func (r *Room) setCanCallLandlordUser(canCallLandlordUser *User) {
	r.canCallLandlordUser = canCallLandlordUser
}

//获取地主出牌次数
func (r *Room) getLandlordPlayCardCount() int {
	return r.landlordPlayCardCount
}

//设置地主出牌次数
func (r *Room) setLandlordPlayCardCount(count int) {
	r.landlordPlayCardCount = count
}

//累加地主出牌次数
func (r *Room) updteLandlordPlayCardCount() {
	r.landlordPlayCardCount += 1
}

//获取农民出牌次数
func (r *Room) getFarmerPlayCardCount() int {
	return r.farmerPlayCardCount
}

//设置农民出牌次数
func (r *Room) setFarmerPlayCardCount(count int) {
	r.farmerPlayCardCount = count
}

//累加农民出牌次数
func (r *Room) updteFarmerPlayCardCount() {
	r.farmerPlayCardCount += 1
}

//获取本局任务
func (r *Room) getCouncilTask() *Task {
	return r.councilTask
}

//设置本局任务
func (r *Room) setCouncilTask(councilTask *Task) {
	r.councilTask = councilTask
}

//获取所有玩家的积分
func (r *Room) getUsersVideoIntegral() []int {
	return r.usersVideoIntegral
}

//获取春天的状态
func (r *Room) getSpringStatus() int {
	return r.springStatus
}

//设置春天的状态
func (r *Room) setSpringStatus(springStatus int) {
	r.springStatus = springStatus
}

//根据userid获取玩家积分
func (r *Room) getUserVideoIntegral(user *User) int {
	userIndex := user.getIndex()
	return r.getUsersVideoIntegral()[userIndex]
}

//根据userid设置玩家积分
func (r *Room) setUserVideoIntegral(user *User, videoIntegral int) {
	userIndex := user.getIndex()
	r.getUsersVideoIntegral()[userIndex] = videoIntegral
}

//重开
func (r *Room) reStart() {
	r.resetUsers()
	r.closeUserCountDown()
	r.SetRoomStatus(RoomStatus_Setout)
	r.reset()
}

//玩家转变成地主
func (r *Room) userTurnLandlord(user *User) {
	logger.Debugf("%s 成为地主", *user.getUID())
	user.setLandlord(true)
	r.setLandlord(user)
	farmers := []*User{}
	for _, u := range r.getUsers() {
		if u != user {
			farmers = append(farmers, u)
		}
	}
	r.setFarmers(farmers)
	r.addCardsToLandlord()
	r.showBaseCards(nil)
	r.openDouble()
}

/*
亮底牌
push:BaseCards_Push,地主userid,cardid$cardid$cardid,底牌类型,底牌倍数,是否加入牌中
*/
func (r *Room) showBaseCards(user *User) {
	if r.getLandlord() == nil {
		return
	}
	//	r.setBaseCards([]Card{Card{Suit: 1, Priority: 1}, Card{Suit: 1, Priority: 2}, Card{Suit: 1, Priority: 3}})
	cards := r.getBaseCards()
	cardsType, multiple := r.getBaseCardsInfo()
	userids := []string{}
	addToCards := 0
	if user == nil { //只执行一次(地主出现的时候)
		//根据底牌加倍
		if multiple > 1 {
			r.setMultiple(r.getMultiple() * multiple)
			r.pushMultiple()
		}
		userids = r.getUserIDs()
		addToCards = 1
	} else { //短线重连进来的
		userids = []string{*user.getUserID()}
	}
	message := fmt.Sprintf("%s,%s,%d,%d,%d", *r.getLandlord().getUserID(), *r.getCardsID(cards), cardsType, multiple, addToCards)
	if user == nil {
		pushMessageToUsers("BaseCards_Push", []string{message}, userids)
		r.pushJudgment("BaseCards_Push", message)
	} else {
		pushMessageToUsers("BaseCards_Push", []string{message}, userids)
	}
}

//将底牌放入地主牌面中
func (r *Room) addCardsToLandlord() {
	cards := r.getBaseCards()
	landlord := r.getLandlord()
	if landlord != nil {
		var tmpCards CardList
		tmpCards = landlord.getCards()
		tmpCards = append(tmpCards, cards...)
		sort.Sort(tmpCards)
		for i := 0; i < len(tmpCards); i++ {
			tmpCards[i].Index = i
		}
		landlord.setCards(tmpCards)
	}
}

/*
获取牌的类型(-1不是特殊底牌 0豹子 1同花 2顺子 3王炸 4同花顺)
*/
func (r *Room) getBaseCardsInfo() (cardsType int, multiple int) {
	cardsType = -1
	multiple = 1
	var cards CardList = r.getBaseCards()
	shunzi := []int{}
	tonghua := map[int]bool{}
	baozi := map[int]bool{}
	wangzha := map[int]bool{}
	for _, card := range cards {
		if card.Priority < Priority_Two {
			if len(shunzi) == 0 {
				shunzi = append(shunzi, card.Priority)
			} else {
				if shunzi[len(shunzi)-1]+1 == card.Priority {
					shunzi = append(shunzi, card.Priority)
				}
			}
		}
		tonghua[card.Suit] = true
		baozi[card.Priority] = true
		if card.Priority >= Priority_SKing {
			wangzha[card.Priority] = true
		}
	}
	isShunzi := len(shunzi) == 3
	isTonghua := len(tonghua) == 1
	isBaozi := len(baozi) == 1
	isWangzha := len(wangzha) == 2
	isTonghuaShun := isShunzi && isTonghua
	if isTonghuaShun {
		cardsType = 4
		multiple = 4
	} else if isWangzha && false {
		cardsType = 3
		multiple = 2
	} else if isShunzi {
		cardsType = 2
		multiple = 2
	} else if isTonghua {
		cardsType = 1
		multiple = 2
	} else if isBaozi {
		cardsType = 0
		multiple = 2
	}
	return cardsType, multiple
}

//获取牌的ID列表
func (u *Room) getCardsID(cards []Card) *string {
	buff := bytes.Buffer{}
	for _, card := range cards {
		buff.WriteString(fmt.Sprintf("%d$", card.ID))
	}
	cardsid := RemoveLastChar(buff)
	return cardsid
}

//获取开赛后的状态
func (r *Room) getMatchingStatus() int {
	return r.matchingStatus
}

//设置开赛后的状态
func (r *Room) setMatchingStatus(matchingStatus int) {
	r.matchingStatus = matchingStatus
}

//获取裁判
func (r *Room) getJudgmentUser() *User {
	return r.judgmentUser
}

//设置裁判
func (r *Room) setJudgmentUser(judgmentUser *User) {
	r.judgmentUser = judgmentUser
}

//获取房间基数
func (r *Room) getCardinality() int {
	return r.cardinality
}

//设置房间基数
func (r *Room) setCardinality(cardinality int) {
	r.cardinality = cardinality
}

//获取房间底分
func (r *Room) getBaseScore() int {
	return r.baseScore
}

//设置房间底分
func (r *Room) setBaseScore(baseScore int) {
	r.baseScore = baseScore
}

/*
推送倍率
push:Multiple_Push,倍数
*/
func (r *Room) pushMultiple() {
	multiple := strconv.Itoa(r.getRealityMultiple())
	pushMessageToUsers("Multiple_Push", []string{multiple}, r.getUserIDs())
	r.pushJudgment("Multiple_Push", multiple)
}

//获取房间倍数
func (r *Room) getMultiple() int {
	return r.multiple
}

//设置房间倍数
func (r *Room) setMultiple(multiple int) {
	r.multiple = multiple
}

//两倍房间倍数并推送
func (r *Room) doubleMultiple() {
	r.setMultiple(r.getMultiple() * 2)
	r.pushMultiple()
}

//三倍房间倍数并推送
func (r *Room) tripleMultiple() {
	r.setMultiple(r.getMultiple() * 3)
	r.pushMultiple()
}

//获取流局倍数
func (r *Room) getLiujuMultiple() int {
	return r.liujuMultiple
}

//设置流局倍数
func (r *Room) setLiujuMultiple(liujuMultiple int) {
	r.liujuMultiple = liujuMultiple
}

//获取房间真实倍数
func (r *Room) getRealityMultiple() int {
	return r.getMultiple() * r.getLiujuMultiple()
}

//更新出牌的轮次
func (r *Room) updatePlayRound() int {
	r.playRound += 1
	return r.playRound
}

//获取出牌的轮次
func (r *Room) getPlayRound() int {
	return r.playRound
}

//设置出牌的轮次
func (r *Room) setPlayRound(playRound int) {
	r.playRound = playRound
}

//更新出牌的次数
func (r *Room) updatePlayTime() int {
	r.playTime += 1
	return r.playTime
}

//获取出牌的次数
func (r *Room) getPlayTime() int {
	return r.playTime
}

//获取出牌的次数
func (r *Room) setPlayTime(playTime int) {
	r.playTime = playTime
}

//获取剩余大王的数量
func (r *Room) getSurplusBKingCount() int {
	return r.surplusBKingCount
}

//设置剩余大王的数量
func (r *Room) setSurplusBKingCount(v int) {
	r.surplusBKingCount = v
}

//更新剩余大王的数量
func (r *Room) updateSurplusBKingCount() {
	r.surplusBKingCount = r.surplusBKingCount - 1
}

//获取剩余小王的数量
func (r *Room) getSurplusSKingCount() int {
	return r.surplusSKingCount
}

//设置剩余小王的数量
func (r *Room) setSurplusSKingCount(v int) {
	r.surplusSKingCount = v
}

//更新剩余小王的数量
func (r *Room) updateSurplusSKingCount() {
	r.surplusSKingCount = r.surplusSKingCount - 1
}

//获取剩余2的数量
func (r *Room) getSurplusTwoCount() int {
	return r.surplusTwoCount
}

//设置剩余2的数量
func (r *Room) setSurplusTwoCount(v int) {
	r.surplusTwoCount = v
}

//更新剩余2的数量
func (r *Room) updateSurplusTwoCount() {
	r.surplusTwoCount = r.surplusTwoCount - 1
}

//获取设置牌权的命令
func (r *Room) getSetCtlMsg() []string {
	return r.setCtlMsg
}

//设置牌权的内容,推送残局时候用
func (r *Room) setSetCtlMsg(setCtlMsg []string) {
	r.setCtlMsg = setCtlMsg
}

//获取初始牌数量是否完整
func (r *Room) initCardCountIsIntegrity() bool {
	return cardCount == perCapitaCardCount
}

//获取房间人数
func (r *Room) GetPCount() int {
	return r.pcount
}

//更新房间人数
func (r *Room) updatePCount(v int) {
	r.pcount = r.pcount + v
}

//获取房间观战人数
func (r *Room) GetIdlePCount() int {
	return len(r.idleusers)
}

//根据index获取玩家
func (r *Room) getUserByIndex(index int) *User {
	return r.users[index]
}

//获取房间入座人数
func (r *Room) getUserCount() int {
	count := 0
	for _, user := range r.users {
		if user != nil {
			count += 1
		}
	}
	return count
}

//获取准备中的玩家数量
func (r *Room) getSetoutCount() int {
	count := 0
	for _, user := range r.users {
		if user != nil {
			if user.getStatus() == UserStatus_Setout {
				count += 1
			}
		}
	}
	return count
}

/*
获取玩家UserID字符串集合
in:是否刷新
*/
func (r *Room) getUserIDs(args ...bool) []string {
	if len(args) > 0 {
		if args[0] {
			r.userids = nil
		}
	}
	if r.userids == nil {
		r.userids = []string{}
		for _, user := range r.users {
			if user != nil {
				r.userids = append(r.userids, *user.userid)
			}
		}
	}
	return r.userids
}

/*
获取未落座玩家UserID字符串集合
in:是否刷新
*/
func (r *Room) getIdleUserIDs(args ...bool) []string {
	if len(args) > 0 {
		if args[0] {
			r.idleuserids = nil
		}
	}
	if r.idleuserids == nil {
		r.idleuserids = []string{}
		for _, user := range r.idleusers {
			if user != nil {
				r.idleuserids = append(r.idleuserids, *user.getUserID())
			}
		}
	}
	return r.idleuserids
}

/*
获取(UserID+IdleUserID)字符串集合
in:是否刷新
*/
func (r *Room) getAllUserIDs() []string {
	userids := r.getUserIDs(true)
	idleuserids := r.getIdleUserIDs(true)
	userids = InsertStringSlice(userids, idleuserids, len(userids))
	return userids
}

//获取比赛类型
func (r *Room) GetMatchID() int {
	return r.matchid
}

//设置比赛类型
func (r *Room) setMatchID(matchID int) {
	r.matchid = matchID
}

//获取总轮次
func (r *Room) getInnings() int {
	return r.innings
}

//设置当前轮次
func (r *Room) setInnings(innings int) {
	r.innings = innings
}

//获取当前轮次
func (r *Room) getInning() int {
	return r.inning
}

//设置当前轮次
func (r *Room) setInning(inning int) {
	r.inning = inning
}

//获取常规赛局数
func (r *Room) getInningRegular() int {
	return r.inningRegular
}

//设置常规赛局数
func (r *Room) setInningRegular(inningRegular int) {
	r.inningRegular = inningRegular
}

//获取房间类型
func (r *Room) GetRoomType() int {
	return r.roomtype
}

//设置房间类型
func (r *Room) setRoomType(roomType int) {
	r.roomtype = roomType
}

//获取牌权玩家
func (r *Room) getControllerUser() *User {
	return r.cuser
}

//设置牌权玩家
func (r *Room) setControllerUser(user *User) {
	r.cuser = user
}

//获取当前牌
func (r *Room) getCurrentCards() []Card {
	return r.cards
}

//设置当前牌
func (r *Room) setCurrentCards(cards []Card) {
	r.cards = cards
}

//获取当前牌的玩家
func (r *Room) getCurrentCardsUser() *User {
	return r.cardsuser
}

//设置当前牌的玩家
func (r *Room) setCurrentCardsUser(user *User) {
	r.cardsuser = user
}

//获取房间状态
func (r *Room) GetRoomStatus() int {
	return r.status
}

//设置房间状态
func (r *Room) SetRoomStatus(status int) {
	r.status = status
}

//获取落座的所有玩家
func (r *Room) getUsers() []*User {
	return r.users
}

//获取未落座的所有玩家
func (r *Room) getIdleUsers() map[string]*User {
	return r.idleusers
}

/*
把房间中所有玩家在负载均衡服务器上的信息都删除
重置玩家
*/
func (r *Room) deleteUsersInfo() {
	users := r.getUsers()
	for _, user := range users {
		if user != nil {
			user.deleteUserInfo()
		}
	}
}

/*
重置房间中所有的玩家
*/
func (r *Room) resetUsers() {
	users := r.getUsers()
	for _, user := range users {
		if user != nil {
			user.reset()
		}
	}
}

//关闭房间
func (r *Room) close() {
	RoomManage.removeRoom(r)
}

//给裁判提送信息
func (r *Room) pushJudgment(funcName string, message string) {
	if judgmentUser := r.getJudgmentUser(); judgmentUser != nil {
		judgmentUser.push(funcName, &message)
	}
}

//设置所有人托管状态
func (r *Room) SetAllUsersTrusteeshipStatus(status bool) {
	for _, user := range r.getUsers() {
		if user != nil {
			user.trusteeship = status
		}
	}
}

/*
所有选手端是否在线
*/
func (r *Room) AllUsersOnlinePush() {
	for _, user := range r.getUsers() {
		if user != nil {
			status := 0
			if user.getOnline() {
				status = 1
			}
			r.pushJudgment("Online_Push", fmt.Sprintf("%s|%d", *user.getUserID(), status))
		}
	}
}
