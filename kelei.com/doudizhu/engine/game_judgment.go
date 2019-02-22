package engine

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	. "kelei.com/utils/common"
)

const (
	DEAL_NORMAL = iota //默认没有好坏牌
	DEAL_RED           //红方牌好
	DEAL_BLUE          //蓝方牌好
)

/*
获取房间列表
*/
func GetRooms(args []string) *string {
	buff := bytes.Buffer{}
	rooms := RoomManage.GetRooms()
	for _, room := range rooms {
		buff.WriteString(*room.GetRoomID())
		buff.WriteString("|")
	}
	res := *RowsBufferToString(buff)
	return &res
}

/*
裁判端进入房间
in:roomid
out:1
push:房间状态、所有人剩余牌、当前轮的出牌信息、当前出牌状态、所有人剩余牌数量
*/
func MatchingJudgment(args []string) *string {
	userid := args[0]
	roomid := args[1]
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	user := UserManage.GetUser(&userid)
	user.setUserType(TYPE_JUDGMENT)
	user.setRoom(room)
	room.setJudgmentUser(user)
	room.matchingPush(nil)
	if room.isMatching() {
		//开赛前玩家操作的推送
		user.userHandle()
		//展示底牌(包含谁是地主)
		room.showBaseCards(user)
		//推送所有人剩余的牌
		pushAllUserSurplusCards(room)
		//推送当前轮的出牌信息
		user.pushCyclePlayCardInfo()
		//推送当前出牌状态
		ctlMsg := room.getSetCtlMsg()
		if len(ctlMsg) > 0 {
			user.setController(ctlMsg[0])
		}
		//暂停状态
		user.pushPauseStatus()
		//所有选手端是否在线
		room.AllUsersOnlinePush()
	}
	return &Res_Succeed
}

/*
推送所有人剩余牌
*/
func pushAllUserSurplusCards(r *Room) {
	users := r.getUsers()
	messages := []string{}
	for _, user := range users {
		buffer := bytes.Buffer{}
		for _, card := range user.cards {
			buffer.WriteString(strconv.Itoa(card.ID))
			buffer.WriteString("|")
		}
		str := buffer.String()
		if str != "" {
			str = str[0 : len(str)-1]
		}
		messages = append(messages, str)
	}
	//给裁判推送所有人的牌局
	r.pushJudgment("Opening_Judgment_Push", strings.Join(messages, "$"))
}

/*
发牌
in:roomid,[牌模式]
out:-1局数已打满
	1成功
*/
func Deal(args []string) *string {
	roomid := args[1]
	mode := DEAL_NORMAL
	if len(args) > 2 {
		mode, _ = strconv.Atoi(args[2])
	}
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	if !room.userEnough() {
		return &Res_Unknown
	}
	if room.getInning() > room.getInnings() {
		res := "-1"
		return &res
	}
	room.reset()
	room.setDealMode(mode)
	room.deal()
	return &Res_Succeed
}

/*
开牌
in:roomid
out:1
*/
func Begin(args []string) *string {
	//	userid := args[0]
	roomid := args[1]
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	if !room.userEnough() {
		return &Res_Unknown
	}
	if room.GetRoomStatus() != RoomStatus_Deal {
		return &Res_Unknown
	}
	go room.begin()
	return &Res_Succeed
}

/*
暂停
in:roomid
out:1
push:[Pause_Push] 1
*/
func Pause(args []string) *string {
	//	userid := args[0]
	roomid := args[1]
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	if !room.userEnough() {
		return &Res_Unknown
	}
	room.pause()
	return &Res_Succeed
}

/*
恢复
in:roomid
out:1
push:[Resume_Push] 1
*/
func Resume(args []string) *string {
	//	userid := args[0]
	roomid := args[1]
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	if !room.userEnough() {
		return &Res_Unknown
	}
	room.resume()
	return &Res_Succeed
}

/*
解散牌局
in:roomid
out:1
push:[Dissolve_Push] 1
*/
func Dissolve(args []string) *string {
	roomid := args[1]
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	if !room.userEnough() {
		return &Res_Unknown
	}
	room.dissolve()
	return &Res_Succeed
}

/*
获取裁判后台数据
in:roomid
out:分值,分值,分值,当前局,总局数,常规赛
*/
func GetJudgmentBgInfo(args []string) *string {
	roomid := args[1]
	room := RoomManage.GetRoom(roomid)
	if room == nil {
		return &Res_Unknown
	}
	videoIntegrals := [3]int{}
	users := room.getUsers()
	for i, user := range users {
		if user != nil {
			videoIntegrals[i] = room.getUserVideoIntegral(user)
		}
	}
	integrals := fmt.Sprintf("%d,%d,%d", videoIntegrals[0], videoIntegrals[1], videoIntegrals[2])
	message := fmt.Sprintf("%s,%d,%d,%d", integrals, room.getInning(), room.getInnings(), room.getInningRegular())
	return &message
}
