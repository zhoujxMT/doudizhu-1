/*
玩家-操作-要不起
*/

package engine

import (
	. "kelei.com/utils/common"
)

//强制过牌(要不起)
func ForceCheckCard(args []string) *string {
	res := Res_Succeed
	userid := args[0]
	user := UserManage.GetUser(&userid)
	//玩家没有操作权限
	if !user.getHandlePerm() {
		return &Res_NoPerm
	}
	CheckCard([]string{userid})
	return &res
}
