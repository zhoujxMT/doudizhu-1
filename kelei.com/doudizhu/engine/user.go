/*
玩家
*/

package engine

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	. "kelei.com/utils/common"
	"kelei.com/utils/delaymsg"
	"kelei.com/utils/frame"
	"kelei.com/utils/logger"
)

//玩家类型
const (
	TYPE_USER     = iota //玩家
	TYPE_JUDGMENT        //裁判
)

//玩家当前状态
const (
	UserStatus_NoSetout  = iota //未准备
	UserStatus_Setout           //已准备
	UserStatus_NoPass           //没过牌
	UserStatus_Pass             //过牌
	UserStatus_NoSitDown        //未落座
)

const (
	TeamMark_Landlord = iota //地主
	TeamMark_Peasant         //农民
)

const (
	MatchResult_Win  = 1  //胜
	MatchResult_Lose = -1 //负
)

const (
	PlayType_Normal  = iota //正常出牌
	PlayType_EndGame        //重新进游戏，残局下，当前轮的出牌信息
)

const (
	EffectType_MatchItem = 2003 //比赛中使用的道具
)

const (
	HANDLE_NOP = iota //没操作
	HANDLE_NO         //否(没叫、没抢、没加倍)
	HANDLE_YES        //是（叫了、抢了、加倍了）
)

//排名基点
var basePoint []int = []int{4, 2, 0, 0, -2, -4, 4}

type User struct {
	uid                *string                //平台userid
	userid             *string                //游戏userid
	conn               net.Conn               //链接
	GateRpc            string                 //网关地址
	sessionKey         string                 //sessionkey
	token              string                 //access_token
	secret             string                 //secret
	headUrl            string                 //头像地址
	room               *Room                  //房间
	status             int                    //玩家状态
	cards              []Card                 //牌列表
	index              int                    //座位编号
	teamMark           int                    //团队标示
	matchResult        int                    //比赛结果(1胜 -1负)
	dm                 *delaymsg.DelayMessage //倒计时
	matchID            int                    //当前所在的比赛类型
	basePoint          int                    //比赛获得基点
	autoTimes          int                    //倒计时结束自动操作的次数
	trusteeship        bool                   //是否托管
	integral           int                    //好友同玩积分,单轮结束清0
	isAI               bool                   //是否是玩家逃跑后,变成的AI
	online             bool                   //玩家是否在线
	chatLastTime       time.Time              //发言的最后时间,用来做冷却
	effectTypes        map[int]int            //使用的道具效果列表
	itemEffect         string                 //道具效果
	userType           int                    //玩家类型(裁判、选手)
	callScore          int                    //叫地主的分值
	openHandStatus     int                    //名牌状态(0没操作 1没明 2明了)
	callLandlordStatus int                    //叫地主状态(0没操作 1没叫 2叫了)
	rushLandlordStatus int                    //抢地主状态(0没操作 1没抢 2抢了)
	doubleStatus       int                    //加倍状态(0没操作 1没加倍 2加倍)
	isLandlord         bool                   //是否是地主
	cardTypeRecord     map[int]int            //打出的牌型记录(map[cardType]count)
	isTaskFinish       bool                   //是否任务完成
	playCardCount      int                    //出牌次数
}

//获取redis中的key
func (u *User) getKey() string {
	return "user:" + *u.getUserID()
}

//转换成参数返回
func (u *User) getArgs() []string {
	return []string{*u.getUserID()}
}

//是否正在比赛
func (u *User) isMatching() bool {
	if u == nil {
		return false
	}
	if u.getRoom() != nil && u.getRoom().isMatching() {
		return true
	}
	return false
}

//获取玩家是否有操作权限
func (u *User) getHandlePerm() bool {
	//不在比赛中,不是活动的,已过牌
	if !u.isMatching() || u.isPass() {
		return false
	}
	return u.currCtlIsSelf()
}

//当前牌权是不是自己
func (u *User) currCtlIsSelf() bool {
	room := u.getRoom()
	//房间当前牌权的玩家
	if room.getControllerUser() == u {
		return true
	}
	return false
}

//获取牌的ID列表
func (u *User) getCardsID() *string {
	buff := bytes.Buffer{}
	for _, card := range u.getCards() {
		buff.WriteString(fmt.Sprintf("%d$", card.ID))
	}
	cardsid := RemoveLastChar(buff)
	return cardsid
}

//是否过牌
func (u *User) isPass() bool {
	return u.getStatus() == UserStatus_Pass
}

//删除玩家手里的牌
func (u *User) updateUserCards(play_indexs []int) {
	userCards := u.getCards()
	for i := len(play_indexs) - 1; i >= 0; i-- {
		index := play_indexs[i]
		userCards = append(userCards[:index], userCards[index+1:]...)
	}
	for k := 0; k < len(userCards); k++ {
		userCards[k].Index = k
	}
	u.setCards(userCards)
}

//重置玩家
func (u *User) reset() {
	u.setMatchResult(MatchResult_Lose)
	u.status = UserStatus_NoSetout
	u.autoTimes = 0
	u.trusteeship = false
	u.cardTypeRecord = map[int]int{}
	u.closeCountDown()
	u.setCallScore(0)
	u.setOpenHandStatus(0)
	u.setCallLandlordStatus(0)
	u.setRushLandlordStatus(0)
	u.setDoubleStatus(0)
	u.setLandlord(false)
	u.setPlayCardCount(0)
	u.setTaskFinish(false)
	u.setIsAI(false)
	u.setOnline(true)
	u.resume()
}

//关闭玩家倒计时
func (u *User) closeCountDown() {
	u.close_countDown_playCard()
	u.close_countDown_setOut()
	u.close_countDown_handle()
}

func (u *User) GetConn() net.Conn {
	return u.conn
}

func (u *User) SetConn(conn net.Conn) {
	u.conn = conn
}

func (u *User) GetGateRpc() string {
	return u.GateRpc
}

func (u *User) SetGateRpc(gateRpc string) {
	u.GateRpc = gateRpc
}

func (u *User) GetSessionKey() string {
	return u.sessionKey
}

func (u *User) SetSessionKey(sessionKey string) {
	u.sessionKey = sessionKey
}

func (u *User) GetToken() string {
	return u.token
}

func (u *User) SetToken(token string) {
	u.token = token
}

func (u *User) GetSecret() string {
	return u.secret
}

func (u *User) SetSecret(secret string) {
	u.secret = secret
}

func (u *User) GetHeadUrl() string {
	return u.headUrl
}

func (u *User) SetHeadUrl(headUrl string) {
	u.headUrl = headUrl
}

func (u *User) getCallScore() int {
	return u.callScore
}

func (u *User) setCallScore(callScore int) {
	u.callScore = callScore
}

func (u *User) getOpenHandStatus() int {
	return u.openHandStatus
}

func (u *User) setOpenHandStatus(openHandStatus int) {
	u.openHandStatus = openHandStatus
}

func (u *User) getCallLandlordStatus() int {
	return u.callLandlordStatus
}

func (u *User) setCallLandlordStatus(callLandlordStatus int) {
	u.callLandlordStatus = callLandlordStatus
}

func (u *User) getRushLandlordStatus() int {
	return u.rushLandlordStatus
}

func (u *User) setRushLandlordStatus(rushLandlordStatus int) {
	u.rushLandlordStatus = rushLandlordStatus
}

func (u *User) getDoubleStatus() int {
	return u.doubleStatus
}

func (u *User) setDoubleStatus(doubleStatus int) {
	u.doubleStatus = doubleStatus
}

func (u *User) getLandlord() bool {
	return u.isLandlord
}

func (u *User) setLandlord(isLandlord bool) {
	u.isLandlord = isLandlord
}

//获取打出此牌型的次数
func (u *User) getCardTypeCount(cardType int) int {
	return u.cardTypeRecord[cardType]
}

//添加打出此牌型的次数
func (u *User) addCardTypeCount(cardType int) {
	u.cardTypeRecord[cardType] += 1
}

func (u *User) getTaskFinish() bool {
	return u.isTaskFinish
}

func (u *User) setTaskFinish(isTaskFinish bool) {
	u.isTaskFinish = isTaskFinish
}

func (u *User) getPlayCardCount() int {
	return u.playCardCount
}

func (u *User) setPlayCardCount(playCardCount int) {
	u.playCardCount = playCardCount
}

func (u *User) updatePlayCardCount() {
	u.playCardCount += 1
}

func (u *User) getItemEffect() string {
	return u.itemEffect
}

func (u *User) setItemEffect(itemEffect string) {
	u.itemEffect = itemEffect
}

func (u *User) getEffectTypes() map[int]int {
	return u.effectTypes
}

func (u *User) setEffectTypes(effectTypes map[int]int) {
	u.effectTypes = effectTypes
}

func (u *User) getUserType() int {
	return u.userType
}

func (u *User) setUserType(userType int) {
	u.userType = userType
}

func (u *User) getBasePoint() int {
	return u.basePoint
}

func (u *User) setBasePoint(basePoint int) {
	u.basePoint = basePoint
}

func (u *User) getTeamMark() int {
	return u.teamMark
}

func (u *User) setTeamMark(teamMake int) {
	u.teamMark = teamMake
}

func (u *User) getIsAI() bool {
	return u.isAI
}

func (u *User) setIsAI(isAI bool) {
	u.isAI = isAI
}

func (u *User) getOnline() bool {
	return u.online
}

func (u *User) setOnline(online bool) {
	u.online = online
}

func (u *User) getChatLastTime() time.Time {
	return u.chatLastTime
}

func (u *User) setChatLastTime() {
	u.chatLastTime = time.Now()
}

func (u *User) getMatchID() int {
	return u.matchID
}

func (u *User) setMatchID(matchID int) {
	u.matchID = matchID
}

func (u *User) getHYTWIntegral() int {
	return u.integral
}

func (u *User) setHYTWIntegral(integral int) {
	u.integral = integral
}

func (u *User) getUserID() *string {
	return u.userid
}

func (u *User) setUserID(userid *string) {
	u.userid = userid
}

func (u *User) getUID() *string {
	return u.uid
}

func (u *User) setUID(uid *string) {
	u.uid = uid
}

func (u *User) setRoom(room *Room) {
	u.room = room
}

func (u *User) getRoom() *Room {
	return u.room
}

func (u *User) getIndex() int {
	return u.index
}

func (u *User) setIndex(index int) {
	u.index = index
}

func (u *User) getStatus() int {
	return u.status
}

func (u *User) setStatus(status int) {
	u.status = status
}

func (u *User) getCards() []Card {
	return u.cards
}

func (u *User) setCards(cards []Card) {
	u.cards = cards
}

func (u *User) getMatchResult() int {
	return u.matchResult
}

func (u *User) setMatchResult(matchResult int) {
	u.matchResult = matchResult
}

//获取托管状态
func (u *User) getTrusteeship() bool {
	return u.trusteeship
}

//是否能压过当前牌
func (u *User) canExcel() bool {
	args := []string{*u.getUserID()}
	indexs := Hint(args)
	if *indexs == "-2" {
		return false
	}
	return true
}

//是否是自己
func (u *User) isSelf(user *User) bool {
	return *u.getUserID() == *user.getUserID()
}

//给一个玩家推送残局
func (u *User) pushEndGame() {
	logger.Debugf("推送残局")
	room := u.getRoom()
	//推送房间的状态信息
	room.matchingPush(u)
	//推送比赛的信息
	u.pushMatchInfo()
	//玩家回来,设置为在线
	u.setOnline(true)
}

//推送比赛的信息
func (u *User) pushMatchInfo() {
	room := u.getRoom()
	//推送此人剩余的牌
	u.pushSurplusCards()
	//所有人剩余的牌数
	u.pushSurplusCardCount()
	//开赛前玩家操作的推送
	u.userHandle()
	//推送明牌玩家的牌
	u.pushOpenHandUserCard()
	//展示底牌(包含谁是地主)
	room.showBaseCards(u)
	//推送当前轮的出牌信息
	u.pushCyclePlayCardInfo()
	//推送托管状态
	u.TG_Push()
	//推送当前出牌状态
	ctlMsg := room.getSetCtlMsg()
	if len(ctlMsg) > 0 {
		u.setController(ctlMsg[0])
	}
	//推送记牌器信息
	u.pushRememberCard()
	//推送暂停状态
	u.pushPauseStatus()
	//裁判端
	if room.getGameRule() == GameRule_Record {
		room.pushJudgment("Online_Push", fmt.Sprintf("%s|%d", *u.getUserID(), 1))
	}
}

//开赛前玩家操作的推送
func (u *User) userHandle() {
	r := u.getRoom()
	if r.GetRoomStatus() == RoomStatus_Handle {
		if r.getLandlord() == nil {
			handleType := HANDLETYPE_CALL
			if r.getGameType() == GAMETYPE_DOUBLE {
				if r.getBaseScore() > 0 {
					handleType = HANDLETYPE_RUSH
				}
			}
			canHandleUser := r.getCanHandleUser()
			message := fmt.Sprintf("%s,%d,%d,%d", *canHandleUser.getUserID(), handleType, r.getBaseScore(), r.getGameType())
			u.push("Handle_Push", &message)
		}
	}
}

/*
推送谁是地主(残局进入的时候)
push:WhoIsLandlord_Push,userid
*/
func (u *User) whoIsLandlord() {
	landlord := u.getRoom().getLandlord()
	if landlord != nil {
		u.push("WhoIsLandlord_Push", landlord.getUserID())
	}
}

//获取下手玩家
func (u *User) getNextUser() *User {
	index := u.getIndex()
	index += 1
	if index >= pcount {
		index = 0
	}
	return u.getRoom().getUsers()[index]
}

//获取下一个可以获得操作权的玩家
func (u *User) getNextCanHandleUser() *User {
	var user *User
	index := u.getIndex()
	users := u.getRoom().getUsers()
	for i := 0; i < 2; i++ {
		index += 1
		if index >= pcount {
			index = 0
		}
		if users[index].getCallLandlordStatus() != HANDLE_NO {
			user = users[index]
			break
		}
	}
	return user
}

//将操作权转移到下手
func (u *User) transferHandleToNextUser() {
	room := u.getRoom()
	if room.GetRoomStatus() == RoomStatus_Liuju {
		return
	}
	if room.getLandlord() == nil {
		user := u.getNextCanHandleUser()
		handleType := HANDLETYPE_CALL
		if room.getGameType() == GAMETYPE_DOUBLE {
			if room.getBaseScore() > 0 {
				handleType = HANDLETYPE_RUSH
			}
		}
		room.setCanHandleUserAndSetCountDown(user, handleType)
	}
}

/*
推送明牌玩家的牌
*/
func (u *User) pushOpenHandUserCard() {
	room := u.getRoom()
	users := room.getUsers()
	for _, user := range users {
		if user.getOpenHandStatus() == HANDLE_YES {
			message := fmt.Sprintf("%s|%s", *user.getUserID(), *user.getCardsID())
			u.push("OpenHand_Push", &message)
		}
	}
}

/*
推送叫地主
*/
func (u *User) pushCallLandlord() {
	room := u.getRoom()
	message := fmt.Sprintf("%s,%d", *u.getUserID(), room.getBaseScore())
	pushMessageToUsers("CallLandlord_Push", []string{message}, room.getUserIDs())
	room.pushJudgment("CallLandlord_Push", message)
}

/*
推送不叫地主
*/
func (u *User) pushNoCallLandlord() {
	room := u.getRoom()
	message := fmt.Sprintf("%s", *u.getUserID())
	pushMessageToUsers("NoCallLandlord_Push", []string{message}, room.getUserIDs())
	room.pushJudgment("NoCallLandlord_Push", message)
}

/*
推送抢地主
*/
func (u *User) pushRushLandlord() {
	room := u.getRoom()
	message := *u.getUserID()
	pushMessageToUsers("RushLandlord_Push", []string{message}, room.getUserIDs())
	room.pushJudgment("RushLandlord_Push", message)
}

/*
推送不抢地主
*/
func (u *User) pushNoRushLandlord() {
	room := u.getRoom()
	message := *u.getUserID()
	pushMessageToUsers("NoRushLandlord_Push", []string{message}, room.getUserIDs())
	room.pushJudgment("NoRushLandlord_Push", message)
}

/*
推送加倍
*/
func (u *User) pushDouble() {
	room := u.getRoom()
	message := *u.getUserID()
	pushMessageToUsers("Double_Push", []string{message}, room.getUserIDs())
	room.pushJudgment("Double_Push", message)
}

/*
推送不加倍
*/
func (u *User) pushNoDouble() {
	room := u.getRoom()
	message := *u.getUserID()
	pushMessageToUsers("NoDouble_Push", []string{message}, room.getUserIDs())
	room.pushJudgment("NoDouble_Push", message)
}

/*
推送暂停状态
*/
func (u *User) pushPauseStatus() {
	room := u.getRoom()
	matchingStatus := strconv.Itoa(room.getMatchingStatus())
	u.push("Pause_Push", &matchingStatus)
}

/*
SCC_Push(所有人剩余的牌数)
push:数量|数量|数量|数量|数量|数量
*/
func (u *User) pushSurplusCardCount() {
	arr := []string{}
	users := u.getRoom().getUsers()
	cardCount := 0
	for _, user := range users {
		if user.getOpenHandStatus() == HANDLE_YES {
			cardCount = 0
		} else {
			cardCount = len(user.getCards())
		}
		arr = append(arr, strconv.Itoa(cardCount))
	}
	message := strings.Join(arr, "|")
	u.push("SCC_Push", &message)
}

/*
RmbCard_Push(推送记牌器信息)
des:1.没有任何的记牌器道具,不推送	2.数量=-1代表没有对应的记牌器道具
out:大王剩余数量|小王剩余数量|2剩余数量
*/
func (u *User) pushRememberCard() {
	room := u.getRoom()
	BKingCount := -1
	SKingCount := -1
	TwoCount := -1
	//有王的记忆
	if u.getKingMemory() != -1 {
		BKingCount = room.getSurplusBKingCount()
		SKingCount = room.getSurplusSKingCount()
	}
	//有2的记忆
	if u.getTwoMemory() != -1 {
		TwoCount = room.getSurplusTwoCount()
	}
	//如果没有任何的记牌器道具
	if BKingCount == -1 && SKingCount == -1 && TwoCount == -1 {
		return
	}
	message := fmt.Sprintf("%d|%d|%d", BKingCount, SKingCount, TwoCount)
	u.push("RmbCard_Push", &message)
}

//断线重连获取比赛信息
func (u *User) Reconnect() *string {
	res := ""
	room := u.getRoom()
	if room == nil {
		return &res
	}
	u.setOnline(true)
	if room.isMatching() {
		//推送比赛的信息
		u.pushMatchInfo()
	} else {
		//推送房间的状态信息
		room.matchingPush(u)
	}
	return &res
}

//推送此人剩余的牌
func (u *User) pushSurplusCards() {
	buffer := bytes.Buffer{}
	for _, card := range u.cards {
		buffer.WriteString(strconv.Itoa(card.ID))
		buffer.WriteString("|")
	}
	str := buffer.String()
	if str != "" {
		message := str[0 : len(str)-1]
		//给此人推送剩余的牌局
		u.push("Opening_Push", &message)
	}
}

//推送当前轮的出牌信息
func (u *User) pushCyclePlayCardInfo() {
	room := u.getRoom()
	users_cards := room.users_cards
	for userid, cardids := range users_cards {
		message := fmt.Sprintf("%s,%s,%d,", userid, cardids, PlayType_EndGame)
		u.push("Play", &message)
	}
}

//暂停倒计时
func (u *User) pause() {
	u.dm.PauseTask()
}

//恢复倒计时
func (u *User) resume() {
	u.dm.ResumeTask()
}

//准备倒计时
func (u *User) countDown_setOut(t time.Duration) {
	if u == nil {
		return
	}
	room := u.getRoom()
	//如果是录像就没有准备倒计时,直接准备
	if room.getGameRule() == GameRule_Record {
		u.setStatus(UserStatus_Setout)
		return
	}
	u.dm.AddTask(time.Now().Add(t), "setOutCountDown", func(args ...interface{}) {
		//网关切换成访问游戏服务器
		u.push("ExitMatch", &Res_Succeed)
		//推送玩家退出了房间
		u.push("ExitRoom_Push", &Res_Succeed)
		//退出房间
		ExitMatch([]string{*u.getUserID()})
	}, nil)
}

//关闭准备倒计时
func (u *User) close_countDown_setOut() {
	u.dm.RemoveTask("setOutCountDown")
}

//操作(叫地主...)倒计时
func (u *User) countDown_handle(t time.Duration) {
	if u == nil {
		return
	}
	u.dm.AddTask(time.Now().Add(t), "handleCountDown", func(args ...interface{}) {
		room := u.getRoom()
		if room.getGameType() == GAMETYPE_REGULAR {
			NoCallLandlord(u.getArgs())
		} else {
			if room.getBaseScore() == 0 {
				NoCallLandlord(u.getArgs())
			} else {
				NoRushLandlord(u.getArgs())
			}
		}
	}, nil)
}

//关闭准备倒计时
func (u *User) close_countDown_handle() {
	u.dm.RemoveTask("handleCountDown")
}

//开启出牌倒计时
func (u *User) countDown_playCard(waitTime int) {
	if u.getTrusteeship() {
		waitTime = 1
	}
	t, _ := time.ParseDuration(strconv.Itoa(waitTime) + "s")
	u.close_countDown_playCard()
	logger.Debugf("开启出牌倒计时:%d", waitTime)
	u.dm.AddTask(time.Now().Add(t), "playCardCountDown", func(args ...interface{}) {
		u.timeEnd(waitTime)
	}, []interface{}{1, 2, 3})
}

//关闭出牌倒计时
func (u *User) close_countDown_playCard() {
	if u != nil {
		u.dm.RemoveTask("playCardCountDown")
	}
}

//出牌倒计时结束,自动操作
func (u *User) timeEnd(waitTime int) {
	room := u.getRoom()
	current_cards := room.getCurrentCards()
	args := []string{*u.getUserID()}
	//托管
	if u.getTrusteeship() {
		u.trusteeshipPlayCard()
	} else {
		//压牌
		if len(current_cards) > 0 {
			CheckCard(args)
		} else {
			//出牌
			u.trusteeshipPlayCard()
		}
		if waitTime >= room.playWaitTime_Long {
			u.trusteeshipHandle()
		}
	}
}

//托管出牌
func (u *User) trusteeshipPlayCard() {
	args := []string{*u.getUserID()}
	indexs := Hint(args)
	if *indexs == "-2" {
		CheckCard(args)
		return
	}
	args = append(args, *indexs)
	PlayCard(args)
}

//托管处理
func (u *User) trusteeshipHandle() {
	u.autoTimes += 1
	//倒计时结束自动操作的次数>=1次,进行托管
	if u.autoTimes >= 1 {
		u.setTrusteeship(true)
	}
}

//设置玩家托管
func (u *User) setTrusteeship(status bool) {
	if u.getRoom().getGameRule() == GameRule_Record {
		return
	}
	if !u.isMatching() {
		return
	}
	if u.trusteeship == status {
		u.TG_Push()
		return
	}
	u.trusteeship = status
	if !u.trusteeship {
		u.close_countDown_playCard()
	}
	u.autoTimes = 0
	u.TG_Push()
}

/*
托管推送
out:托管状态(0不托1托)
*/
func (u *User) TG_Push() {
	status := u.trusteeship
	//托管推送
	funcName := "TG_Push"
	message := "0"
	//托管
	if status {
		message = "1"
		room := u.getRoom()
		//如果玩家是当前控牌人,关闭倒计时,玩家立即出牌
		if room.getControllerUser() == u {
			u.close_countDown_playCard()
			u.timeEnd(0)
		}
	}
	u.push(funcName, &message)
}

func (u *User) setController(message string) {
	pushMessageToUsers("SetController_Push", []string{message}, []string{*u.getUserID()})
}

func (u *User) setControllerUsers(message string) {
	pushMessageToUsers("SetController_Push", []string{message}, u.getRoom().getUserIDs())
	u.getRoom().pushJudgment("SetController_Push", message)
}

func (u *User) setCtlUsers(userid string, userstatus string, setControllerStatus int) {
	message := fmt.Sprintf("%s|%s|,,%d,", userid, userstatus, setControllerStatus)
	u.setControllerUsers(message)
}

//给此玩家推送信息
func (u *User) push(funcName string, message *string) {
	if frame.GetMode() == frame.MODE_RELEASE {
		defer func() {
			if p := recover(); p != nil {
				errInfo := fmt.Sprintf("push : { %v }", p)
				logger.Errorf(errInfo)
			}
		}()
	}
	time.Sleep(time.Millisecond)
	conn := u.GetConn()
	xServer := frame.GetRpcxServer()
	msg := fmt.Sprintf("%s&%s&%s", *u.getUserID(), funcName, *message)
	logger.Debugf("推送数据:%s", msg)
	err := xServer.SendMessage(conn, "service_path", "service_method", nil, []byte(msg))
	logger.CheckError(err, fmt.Sprintf("failed to send messsage to %s : ", conn.RemoteAddr().String()))
}

//玩家关闭连接
func (u *User) close() {
	room := u.getRoom()
	if room == nil {
		return
	}
	//没开赛的时候,退出房间
	if !room.isMatching() {
		if u.getStatus() == UserStatus_NoSetout || u.getStatus() == UserStatus_Setout {
			u.exitRoom()
			UserManage.RemoveUser(u)
		}
	} else {
		//开赛后,玩家托管,并设置玩家离线
		u.setOnline(false)
		room.checkMatchPause(u)
	}
}

//玩家进入房间
func (u *User) enterRoom(room *Room) {
	u.room = room
	room.idleusers[*u.getUserID()] = u
}

//玩家“手动找位置”坐下
func (u *User) sitDown(seatIndex int) {
	room := u.getRoom()
	//从未落座的玩家中删除
	delete(room.getIdleUsers(), *u.getUserID())
	//落座
	if room.users[seatIndex] == nil {
		u.setIndex(seatIndex)
		room.users[seatIndex] = u
		room.updatePCount(1)
		u.setStatus(UserStatus_NoSetout)
	}
}

//玩家“自动找位置”坐下
func (u *User) sitDownAuto() {
	room := u.getRoom()
	for i, user := range room.getUsers() {
		if user == nil {
			u.sitDown(i)
			break
		} else {
			//人已经在房间中
			if *user.getUserID() == *u.getUserID() {
				break
			}
		}
	}
}

//玩家对号入座
func (u *User) sitDownPigeon(seatIndex int) {
	room := u.getRoom()
	for i, user := range room.getUsers() {
		if i == seatIndex {
			if user == nil {
				u.sitDown(i)
				break
			} else {
				//人已经在房间中
				if *user.getUserID() == *u.getUserID() {
					break
				}
			}
		}
	}
}

//玩家站起
func (u *User) standUp() {
	if u.getStatus() == UserStatus_NoSitDown {
		return
	}
	room := u.getRoom()
	//从已落座的玩家中删除
	room.users[u.getIndex()] = nil
	//放入未落座区域
	room.getIdleUsers()[*u.getUserID()] = u
	room.updatePCount(-1)
	u.setStatus(UserStatus_NoSitDown)
}

//玩家退出房间
func (u *User) exitRoom() {
	room := u.getRoom()
	if room == nil {
		return
	}
	if u.getStatus() == UserStatus_NoSitDown {
		delete(room.idleusers, *u.getUserID())
	} else {
		room.users[u.getIndex()] = nil
		room.updatePCount(-1)
	}
	totalPCount := room.GetIdlePCount() + room.GetPCount()
	if room.getGameRule() != GameRule_Record && totalPCount <= 0 {
		room.close()
	} else {
		u.exitRoomHandle()
	}
	u.room = nil
	u.reset()
}

//玩家退出房间的处理
func (u *User) exitRoomHandle() {
	room := u.getRoom()
	//好友同玩有人退出房间,房间重置
	if room.GetMatchID() == Match_HYTW {
		room.resetHYTW()
	}
	//房间的状态信息推送
	if room.GetMatchID() == Match_HXS {
		room.matchingHXSPush(nil)
	} else {
		room.matchingPush(nil)
	}
}

//退出正在比赛的房间
func (u *User) exitMatchRoom() {
	room := u.getRoom()
	//不能删！！！
	if room == nil {
		return
	}
	if room.getGameRule() == GameRule_Record {
		return
	}
	//增加逃跑次数
	u.updateFlee(1)
	//获取逃跑扣除的元宝和积分
	fleeCost := *FleeCost([]string{*u.getUserID()})
	arr := strings.Split(fleeCost, "|")
	ingot, _ := strconv.Atoi(arr[0])
	integral, _ := strconv.Atoi(arr[1])
	//扣元宝
	u.updateIngot(-ingot, 5)
	//扣积分
	u.updateIntegral(-integral)
	//删除玩家数据库中的记录信息
	u.deleteUserInfo()
	//将玩家改为AI,释放此玩家
	u.selfToAI()
}

//将玩家改为AI,释放此玩家
func (u *User) selfToAI() {
	u.setIsAI(true)
	u.setTrusteeship(true)
	u.updateUserID()
	UserManage.AddTheUser(u)
	fmt.Println(*u.getUserID(), "逃跑")
}

/*
UpdateUserID_Push
push:需要替换的UserID|新的UserID
*/
func (u *User) updateUserID() {
	room := u.getRoom()
	aiUserID := fmt.Sprintf("%s%s", *room.GetRoomID(), *u.getUserID())
	pushMessageToUsers("UpdateUserID_Push", []string{fmt.Sprintf("%s|%s", *u.getUserID(), aiUserID)}, room.getUserIDs())
	u.setUserID(&aiUserID)
	room.getUserIDs(true)
}

//比赛中对其它玩家使用道具
func (u *User) MatchUseItem(dstuserid string, itemid int) *string {
	user := UserManage.GetUser(&dstuserid)
	if user == nil {
		return &Res_Unknown
	}
	res := u.useItemByItemID(itemid, 1, user.getRealityUserID())
	res = fmt.Sprintf("%s,%d", strings.Split(res, ",")[0], itemid)
	pushMessageToUsers("MUI_Push", []string{fmt.Sprintf("%s,%s,%d", *u.getUserID(), dstuserid, itemid)}, u.getRoom().getUserIDs())
	return &res
}
