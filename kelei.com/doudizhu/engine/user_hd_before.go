/*
玩家-操作-明牌
*/

package engine

import (
	"fmt"
	"strconv"

	. "kelei.com/utils/common"
)

/*
明牌
in:
out:1成功
push:
	1.OpenHand_Push,userid|cardids
	2.Multiple_Push,倍率
*/
func OpenHand(args []string) *string {
	res := Res_Succeed
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil {
		return &Res_Unknown
	}
	room := user.getRoom()
	if room == nil {
		return &Res_Unknown
	}
	if !room.isHandling() {
		return &Res_Unknown
	}
	if user.getOpenHandStatus() != HANDLE_NOP {
		return &Res_Unknown
	}
	user.setOpenHandStatus(HANDLE_YES)
	room.doubleMultiple()
	message := fmt.Sprintf("%s|%s", userid, *user.getCardsID())
	pushMessageToUsers("OpenHand_Push", []string{message}, room.getUserIDs())
	room.pushJudgment("OpenHand_Push", message)
	return &res
}

/*
不明牌
in:
out:1成功
*/
func NoOpenHand(args []string) *string {
	res := Res_Succeed
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil {
		return &Res_Unknown
	}
	room := user.getRoom()
	if room == nil {
		return &Res_Unknown
	}
	if !room.isHandling() {
		return &Res_Unknown
	}
	if user.getOpenHandStatus() != HANDLE_NOP {
		return &Res_Unknown
	}
	user.setOpenHandStatus(HANDLE_NO)
	return &res
}

/*
叫地主
in:分值
out:-1分值无效
	1成功
push:
	1. CallLandlord_Push,userid,底分
	2. BaseCards_Push,地主userid,cardid$cardid$cardid,底牌类型,底牌倍数
des:
	分值(常规赛1-3,加倍赛不用传分值)
	底牌类型(-1无 0顺子 1同花 2豹子 3王炸 4同花顺)
*/
func CallLandlord(args []string) *string {
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil {
		return &Res_Unknown
	}
	room := user.getRoom()
	if room == nil {
		return &Res_Unknown
	}
	if !room.isHandling() {
		return &Res_Unknown
	}
	if user.getCallLandlordStatus() != HANDLE_NOP {
		return &Res_Unknown
	}
	if room.getCanHandleUser() != user {
		return &Res_Unknown
	}
	if room.getLandlord() != nil {
		return &Res_Unknown
	}
	user.close_countDown_handle()
	score := 0
	if room.getGameType() == GAMETYPE_REGULAR { //常规赛
		score, _ = strconv.Atoi(args[1])
		res := "-1"
		if !(score >= 1 && score <= 3) {
			return &res
		}
		if score <= room.getBaseScore() {
			return &res
		}
		//3分直接成为地主
		if score == 3 {
			for _, user := range room.getUsers() {
				user.setCallLandlordStatus(HANDLE_YES)
			}
		}
	} else { //加倍赛
		//有人叫过,不能再叫了
		if room.getBaseScore() > 0 {
			return &Res_Unknown
		}
		score = 3
	}
	room.setBaseScore(score)
	user.setCallScore(score)
	user.setCallLandlordStatus(HANDLE_YES)
	user.pushCallLandlord()
	setLandlordByScore(room)
	user.transferHandleToNextUser()
	return &Res_Succeed
}

/*
不叫地主
in:
out:1成功
push:
	1. NoCallLandlord_Push,userid
	2. BaseCards_Push,地主userid,cardid$cardid$cardid,底牌类型,底牌倍数
des:
	底牌类型(0顺子 1同花 2豹子 3王炸 4同花顺)
*/
func NoCallLandlord(args []string) *string {
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil {
		return &Res_Unknown
	}
	room := user.getRoom()
	if room == nil {
		return &Res_Unknown
	}
	if !room.isHandling() {
		return &Res_Unknown
	}
	if user.getCallLandlordStatus() != HANDLE_NOP {
		return &Res_Unknown
	}
	if room.getGameType() == GAMETYPE_DOUBLE && room.getBaseScore() > 0 {
		return &Res_Unknown
	}
	if room.getCanHandleUser() != user {
		return &Res_Unknown
	}
	if room.getLandlord() != nil {
		return &Res_Unknown
	}
	user.close_countDown_handle()
	user.setCallLandlordStatus(HANDLE_NO)
	user.pushNoCallLandlord()
	setLandlordByScore(room)
	user.transferHandleToNextUser()
	return &Res_Succeed
}

//根据叫的分值设置地主(常规赛)
func setLandlordByScore(room *Room) {
	var user *User
	noHandleUserCount := 0
	for _, u := range room.getUsers() {
		if u.getCallLandlordStatus() == HANDLE_YES {
			if user == nil {
				user = u
			} else {
				if u.getCallScore() > user.getCallScore() {
					user = u
				}
			}
		} else if u.getCallLandlordStatus() == HANDLE_NOP {
			noHandleUserCount += 1
		}
	}
	if noHandleUserCount == 0 {
		if user == nil {
			if room.getGameType() == GAMETYPE_DOUBLE {
				room.reStart()
				room.setLiujuMultiple(room.getLiujuMultiple() * 2)
				room.pushMultiple()
				fmt.Println("流局倍数:", room.getLiujuMultiple())
			}
			liuju := func() {
				room.SetRoomStatus(RoomStatus_Liuju)
				room.getUsers()[0].setCtlUsers("", "", SetController_Liuju)
				room.setMatchingStatus(MatchingStatus_Over)
			}
			liuju()
			if room.getGameRule() != GameRule_Record {
				//重开牌局
				room.reStart()
				//发牌
				room.match_Opening()
				if room.getGameType() == GAMETYPE_DOUBLE {
					//推送倍数
					room.pushMultiple()
				}
			}
		} else {
			room.userTurnLandlord(user)
		}
	}
}

/*
抢地主
in:
out:1成功
push:
	1. RushLandlord_Push,userid
	2. BaseCards_Push,地主userid,cardid$cardid$cardid,底牌类型,底牌倍数
des:
	底牌类型(0顺子 1同花 2豹子 3王炸 4同花顺)
*/
func RushLandlord(args []string) *string {
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil {
		return &Res_Unknown
	}
	room := user.getRoom()
	if room == nil {
		return &Res_Unknown
	}
	if room.getGameType() != GAMETYPE_DOUBLE {
		return &Res_Unknown
	}
	if !room.isHandling() {
		return &Res_Unknown
	}
	if room.getBaseScore() == 0 {
		return &Res_Unknown
	}
	if user.getRushLandlordStatus() != HANDLE_NOP {
		return &Res_Unknown
	}
	if room.getCanHandleUser() != user {
		return &Res_Unknown
	}
	if room.getLandlord() != nil {
		return &Res_Unknown
	}
	user.close_countDown_handle()
	user.setRushLandlordStatus(HANDLE_YES)
	user.pushRushLandlord()
	room.doubleMultiple()
	setLandlordByRush(room)
	user.transferHandleToNextUser()
	return &Res_Succeed
}

/*
不抢地主
in:
out:1成功
push:
	1. NoRushLandlord_Push,userid
	2. BaseCards_Push,地主userid,cardid$cardid$cardid,底牌类型,底牌倍数
des:
	底牌类型(0顺子 1同花 2豹子 3王炸 4同花顺)
*/
func NoRushLandlord(args []string) *string {
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil {
		return &Res_Unknown
	}
	room := user.getRoom()
	if room == nil {
		return &Res_Unknown
	}
	if room.getGameType() != GAMETYPE_DOUBLE {
		return &Res_Unknown
	}
	if !room.isHandling() {
		return &Res_Unknown
	}
	if room.getBaseScore() == 0 {
		return &Res_Unknown
	}
	if user.getRushLandlordStatus() != HANDLE_NOP {
		return &Res_Unknown
	}
	if room.getCanHandleUser() != user {
		return &Res_Unknown
	}
	if room.getLandlord() != nil {
		return &Res_Unknown
	}
	user.close_countDown_handle()
	user.setRushLandlordStatus(HANDLE_NO)
	user.pushNoRushLandlord()
	setLandlordByRush(room)
	user.transferHandleToNextUser()
	return &Res_Succeed
}

//根据抢设置地主(加倍赛)
func setLandlordByRush(room *Room) {
	var user *User
	var callUser *User         //叫地主的玩家
	noHandleUsers := []*User{} //没操作的玩家
	noCallUsers := []*User{}   //不叫地主的玩家
	noRushUsers := []*User{}   //不抢地主的玩家
	for _, u := range room.getUsers() {
		if u.getCallLandlordStatus() == HANDLE_NOP && u.getRushLandlordStatus() == HANDLE_NOP {
			noHandleUsers = append(noHandleUsers, u)
		} else {
			if u.getCallLandlordStatus() == HANDLE_YES {
				callUser = u
			} else if u.getCallLandlordStatus() == HANDLE_NO {
				noCallUsers = append(noCallUsers, u)
			} else if u.getRushLandlordStatus() == HANDLE_NO {
				noRushUsers = append(noRushUsers, u)
			}
		}
	}
	if len(noHandleUsers) == 0 {
		//都不叫 || 都不抢 || 一个不叫一个不抢
		if len(noCallUsers)+len(noRushUsers) == pcount-1 {
			user = callUser
		} else {
			//叫地主的玩家，操作过抢或不抢了
			if callUser.getRushLandlordStatus() != HANDLE_NOP {
				//叫地主的玩家又抢地主了
				if callUser.getRushLandlordStatus() == HANDLE_YES {
					user = callUser
				} else {
					//叫地主的玩家不抢地主,地主给下一个抢地主的玩家
					nextUser := callUser
					for i := 0; i < pcount-1; i++ {
						nextUser = nextUser.getNextCanHandleUser()
						if nextUser.getRushLandlordStatus() == HANDLE_YES {
							user = nextUser
							break
						}
					}
				}
			}
		}
		if user != nil {
			room.userTurnLandlord(user)
		}
	}
}

/*
加倍
in:
out:1成功
push:Double_Push,userid
*/
func Double(args []string) *string {
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil {
		return &Res_Unknown
	}
	room := user.getRoom()
	if room == nil {
		return &Res_Unknown
	}
	if room.getGameType() != GAMETYPE_DOUBLE {
		return &Res_Unknown
	}
	if !room.isHandling() {
		return &Res_Unknown
	}
	if user.getDoubleStatus() != HANDLE_NOP {
		return &Res_Unknown
	}
	user.setDoubleStatus(HANDLE_YES)
	user.pushDouble()
	return &Res_Succeed
}

/*
不加倍
in:
out:1成功
push:NoDouble_Push,userid
*/
func NoDouble(args []string) *string {
	userid := args[0]
	user := UserManage.GetUser(&userid)
	if user == nil {
		return &Res_Unknown
	}
	room := user.getRoom()
	if room == nil {
		return &Res_Unknown
	}
	if room.getGameType() != GAMETYPE_DOUBLE {
		return &Res_Unknown
	}
	if !room.isHandling() {
		return &Res_Unknown
	}
	if user.getDoubleStatus() != HANDLE_NOP {
		return &Res_Unknown
	}
	user.setDoubleStatus(HANDLE_NO)
	user.pushNoDouble()
	return &Res_Succeed
}
