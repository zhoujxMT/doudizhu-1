/*
房间比赛结束
*/

package engine

import (
	"bytes"
	"fmt"
	"time"

	"kelei.com/utils/logger"
)

//比赛结束的检测
func (r *Room) checkMatchingOver() {
	//比赛是否结束
	matchOver := false
	//是否是地主走科
	landlordGo := false
	for _, user := range r.getUsers() {
		//有人走科
		if len(user.getCards()) <= 0 {
			//走科的人是地主
			if user.getLandlord() {
				landlordGo = true
			}
			matchOver = true
		}
	}
	//比赛结束
	if matchOver {
		if landlordGo { //是地主走科
			r.getLandlord().setMatchResult(MatchResult_Win)
			farmers := r.getFarmers()
			for _, farmer := range farmers {
				farmer.setMatchResult(MatchResult_Lose)
			}
		} else { //是农民走科
			r.getLandlord().setMatchResult(MatchResult_Lose)
			farmers := r.getFarmers()
			for _, farmer := range farmers {
				farmer.setMatchResult(MatchResult_Win)
			}
		}
		//比赛结束处理
		r.matchingOverHandle()
	}
}

/*
比赛结束的处理
*/
func (r *Room) matchingOverHandle() {
	//任务奖励
	r.taskAward()
	//春天和反春的判断
	r.springHandle()
	//关闭玩家的倒计时
	r.closeUserCountDown()
	//删除所有玩家的负载均衡服务器信息
	r.deleteUsersInfo()
	//局数增加
	r.inningIncrease()
	//常规赛和加倍赛的切换
	r.gameTypeSwitch()
	//展示剩余排面
	r.showSurplusCards()
	//推送比赛结束的信息给所有玩家
	r.pushMatchingEndInfo()
	//重置加倍场流局倍数
	r.setLiujuMultiple(1)
	//修改网关的比赛状态
	r.updateGateMatchStatus()
	//海选赛
	if r.GetMatchID() == Match_HXS {
		//将所有人踢出房间
		r.removeAllUsers()
	} else {
		r.reset()
		//设置房间为准备中的状态
		r.SetRoomStatus(RoomStatus_Setout)
		//重置所有的玩家
		r.resetUsers()
		//删除离线和逃跑的玩家
		r.removeUsers()
		//推送房间的状态,所有人变成了未准备的状态
		r.matchingPush(nil)
		//开启准备倒计时
		r.setOutCountDown()
	}
	//上报用户数据
	r.reportedUserInfo()
}

//局数增加
func (r *Room) inningIncrease() {
	r.setInning(r.getInning() + 1)
}

//常规赛和加倍赛的切换
func (r *Room) gameTypeSwitch() {
	//当前局数>常规赛局数
	if r.getInning() > r.getInningRegular() {
		r.setGameType(GAMETYPE_DOUBLE)
	} else {
		r.setGameType(GAMETYPE_REGULAR)
	}
}

/*
任务奖励
push:TaskAward_Push,userid,积分数量
*/
func (r *Room) taskAward() {
	for _, user := range r.getUsers() {
		//胜利 && 任务完成
		if user.getMatchResult() == MatchResult_Win && user.getTaskFinish() {
			//发放奖励
			task := r.getCouncilTask()
			logger.Debugf("%s 完成任务《%s》获得奖励%d", *user.getUserID(), task.getContent(), task.getAward())
			r.setUserVideoIntegral(user, r.getUserVideoIntegral(user)+task.getAward())
			message := fmt.Sprintf("%s,%d", *user.getUserID(), task.getAward())
			pushMessageToUsers("TaskAward_Push", []string{message}, r.getUserIDs())
			r.pushJudgment("TaskAward_Push", message)
		}
	}
}

//春天和反春的处理
func (r *Room) springHandle() {
	r.setSpringStatus(0)
	landlordPlayCardCount := r.getLandlordPlayCardCount()
	parmerPlayCardCount := r.getFarmerPlayCardCount()
	//春天
	isSpring := landlordPlayCardCount >= 0 && parmerPlayCardCount <= 0
	//反春
	isReverseSpring := landlordPlayCardCount <= 1 && parmerPlayCardCount > 0
	if isSpring || isReverseSpring {
		r.doubleMultiple()
		if isSpring {
			r.setSpringStatus(1)
		} else {
			r.setSpringStatus(2)
		}
	}
}

//修改网关的比赛状态
func (r *Room) updateGateMatchStatus() {
	users := r.getUsers()
	for _, user := range users {
		if user != nil {
			res := ""
			user.push("MatchEnd", &res)
		}
	}
}

//将所有人踢出房间
func (r *Room) removeAllUsers() {
	users := r.getUsers()
	for _, user := range users {
		if user != nil {
			user.exitRoom()
		}
	}
}

//上报用户数据
func (r *Room) reportedUserInfo() {
	users := r.getUsers()
	for _, user := range users {
		if user != nil && !user.getIsAI() {
			user.uploadWX()
		}
	}
}

//关闭所有玩家的定时器
func (r *Room) closeUserCountDown() {
	for _, user := range r.getUsers() {
		user.closeCountDown()
	}
}

//删除离线和逃跑的玩家
func (r *Room) removeUsers() {
	users := r.getUsers()
	for _, user := range users {
		if user != nil {
			//玩家不在线
			if !user.getOnline() || user.getIsAI() {
				//退出房间
				ExitMatch([]string{*user.getUserID()})
			}
		}
	}
}

/*
ShowSurplusCards_Push(展示剩余牌面)
push:44|,46|1$1$1$1$16$2
*/
func (r *Room) showSurplusCards() {
	buff := bytes.Buffer{}
	users := r.getUsers()
	for _, user := range users {
		buff.WriteString(fmt.Sprintf("%s|%s,", *user.getUserID(), *user.getCardsID()))
	}
	message := buff.String()
	message = message[:len(message)-1]
	time.Sleep(time.Second)
	pushMessageToUsers("ShowSurplusCards_Push", []string{message}, r.getUserIDs())
	time.Sleep(time.Millisecond * 1500)
}

//结算积分信息
var IntegralData_HYTW = []int{40, 20, 0, 0, -20, -40, 40}

//推送比赛结束的信息-经典
func (r *Room) pushMatchingEndInfo_JD() {
	message := ""
	commonScore := r.getCardinality() * r.getBaseScore() * r.getRealityMultiple()
	landlord := r.getLandlord()
	landlordDouble := landlord.getDoubleStatus()
	farmers := r.getFarmers()
	farmer1, farmer2 := farmers[0], farmers[1]
	farmer1Double := farmer1.getDoubleStatus()
	farmer2Double := farmer2.getDoubleStatus()
	//计算分值
	farmer1Score := commonScore * landlordDouble * farmer1Double * farmer1.getMatchResult()
	farmer2Score := commonScore * landlordDouble * farmer2Double * farmer2.getMatchResult()
	landlordScore := -(farmer1Score + farmer2Score)
	//添加分值
	r.setUserVideoIntegral(farmer1, r.getUserVideoIntegral(farmer1)+farmer1Score)
	r.setUserVideoIntegral(farmer2, r.getUserVideoIntegral(farmer2)+farmer2Score)
	r.setUserVideoIntegral(landlord, r.getUserVideoIntegral(landlord)+landlordScore)
	//组合返回字符串
	landlordMess := fmt.Sprintf("%s|%d|%d", *landlord.getUserID(), landlordScore, r.getUserVideoIntegral(landlord))
	farmer1Mess := fmt.Sprintf("%s|%d|%d", *farmer1.getUserID(), farmer1Score, r.getUserVideoIntegral(farmer1))
	farmer2Mess := fmt.Sprintf("%s|%d|%d", *farmer2.getUserID(), farmer2Score, r.getUserVideoIntegral(farmer2))
	message = fmt.Sprintf("%d,%s$%s$%s,%d", landlord.getMatchResult(), landlordMess, farmer1Mess, farmer2Mess, r.getSpringStatus())
	/*
		比赛结束推送
		push:MatchingEnd_Push,地主是否胜利,userid|本场积分|总积分$userid|本场积分|总积分$userid|本场积分|总积分,春天的状态
		des:地主是否胜利(1胜利 -1失败)
			春天的状态(0无1春天2反春)
	*/
	pushMessageToUsers("MatchingEnd_Push", []string{message}, r.getUserIDs())
	r.pushJudgment("MatchingEnd_Push", message)
}

//推送比赛结束的信息-好友同玩
func (r *Room) pushMatchingEndInfo_HYTW() {

}

//好友同玩-重置房间
func (r *Room) resetHYTW() {
	//设置房间为第一局
	r.setInning(1)
	//清空玩家积分
	users := r.getUsers()
	for _, user := range users {
		if user != nil {
			user.setHYTWIntegral(0)
		}
	}
}

//推送比赛结束的信息-海选赛
func (r *Room) pushMatchingEndInfo_HXS() {

}

/*
MatchingEnd_Push(推送比赛结束的信息)
out:
	够级英雄: 等级|积分|经验|当前级别经验|当前级别升级经验|房费|比赛结果(1胜0平-1负)|是否升级|金币,userid$获得元宝$是否破产$buff道具列表$红蓝队$玩家等级|
	好友同玩: 等级|积分|第几局|是否是最后一局|红队积分|蓝队积分,userid$获得积分$是否胜利$红蓝队$玩家等级|
*/
func (r *Room) pushMatchingEndInfo() {
	if r.GetMatchID() == Match_JD {
		r.pushMatchingEndInfo_JD()
	} else if r.GetMatchID() == Match_HYTW {
		r.pushMatchingEndInfo_HYTW()
	} else if r.GetMatchID() == Match_HXS {
		r.pushMatchingEndInfo_HXS()
	}
}
