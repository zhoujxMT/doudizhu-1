/*
玩家-操作-过牌
*/

package engine

import (
	. "kelei.com/utils/common"
)

/*
过牌
in:
out:-1
*/
func CheckCard(args []string) *string {
	res := Res_Succeed
	userid := args[0]
	user := UserManage.GetUser(&userid)
	//玩家没有操作权限
	if !user.getHandlePerm() {
		return &Res_NoPerm
	}
	user.close_countDown_playCard()
	room := user.getRoom()
	currentCardsUser := room.getCurrentCardsUser()
	//新一轮出牌,没有当前牌的玩家(正常情况下不可能出现)
	if currentCardsUser == nil {
		return &Res_Unknown
	}
	//设置玩家过牌
	user.setStatus(UserStatus_Pass)
	//获取哪个玩家获取牌权
	var nextUser = room.getNextUser()
	//更换牌权
	if nextUser != nil {
		room.setController(nextUser, SetController_Press)
	} else {
		//新一轮
		room.setController(currentCardsUser, SetController_NewCycle)
	}
	return &res
}
