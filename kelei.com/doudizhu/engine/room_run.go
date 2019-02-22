/*
房间比赛中
*/

package engine

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"

	"kelei.com/utils/logger"
)

//获取牌的ID列表
func getCardIDs(cards []Card) *string {
	cardids := bytes.Buffer{}
	for _, card := range cards {
		cardids.WriteString(strconv.Itoa(card.ID))
		cardids.WriteString("|")
	}
	cardids_s := cardids.String()
	cardids_s = cardids_s[:len(cardids_s)-1]
	return &cardids_s
}

/*
设置牌权（真实牌权、等待牌权）
users是玩家：
	说明只有一个人获得 “真实牌权” ,给所有玩家推送的信息都是一样的
users是玩家数组：
	说明是多个人获得 “等待牌权” ,给所有玩家推送的信息是不一样的,这种情况只有两种可能
	1.级牌出牌之后,多人都可以（烧牌）
	2.（级牌让牌）之后,多人都可以上牌
*/
func (r *Room) setController(user *User, status int) {
	//只有最后一个玩家点出牌,才会找不到下一个出牌人
	if user == nil {
		logger.Debugf("比赛结束")
		return
	}
	//获取玩家的操作
	userHandle := r.getUserHandle()
	//设置牌权
	r.setControllerUser(user)
	if status == SetController_NewCycle {
		//新一轮
		r.newCycle()
		//更新轮次
		r.updatePlayRound()
	}
	//等待时间
	waitTime := r.getWaitTime(status)
	//给玩家添加操作倒计时
	user.countDown_playCard(waitTime)
	message := fmt.Sprintf("%s,%s,%d,%d", userHandle, *user.getUserID(), status, waitTime)
	r.setSetCtlMsg([]string{message})
	pushMessageToUsers("SetController_Push", []string{message}, r.getUserIDs())
	r.pushJudgment("SetController_Push", message)
}

//牌面玩家占牌,新一轮出牌
func (r *Room) currentCardsUserNewCycle() {
	time.Sleep(time.Millisecond * 1500)
	currentCardsUser := r.getCurrentCardsUser()
	r.setController(currentCardsUser, SetController_NewCycle)
}

//获取操作倒计时
func (r *Room) getWaitTime(status int) int {
	waitTime := r.playWaitTime_Long
	if status == SetController_Pass {
		waitTime = r.playWaitTime
	}
	return waitTime
}

//获取玩家操作的数据
func (r *Room) getUserHandle() string {
	userHandle := ""
	user := r.getControllerUser()
	if user == nil {
		return userHandle
	}
	status := user.getStatus()
	//如果控制牌权的人的状态是
	if status == UserStatus_Pass {
		info := ""
		userHandle = fmt.Sprintf("%s|%d|%s", *user.getUserID(), status, info)
	}
	return userHandle
}

//获取下一顺位出牌人
func (r *Room) getNextUser() *User {
	controllerUser := r.getControllerUser()
	currentCardsUser := r.getCurrentCardsUser()
	users := r.getUsers()
	index := controllerUser.getIndex()
	var user *User
	for i := 0; i < pcount-1; i++ {
		index = getNextUserIndex(index)
		u := users[index]
		//不是出牌人
		if u != currentCardsUser {
			//没过牌
			if u.getStatus() == UserStatus_NoPass {
				user = u
				break
			}
		}
	}
	return user
}

//获取下一个玩家index
func getNextUserIndex(index int) int {
	index = index + 1
	if index > pcount-1 {
		index = 0
	}
	return index
}

//根据优先级获取牌列表
func getCardsByPriority(cs []Card, priority int) []Card {
	cards := []Card{}
	for _, card := range cs {
		if card.Priority < priority {
			break
		}
		if card.Priority == priority {
			cards = append(cards, card)
		}
	}
	return cards
}

//根据优先级获取牌是否存在
func isExistByPriority(cards []Card, priority int) bool {
	exist := false
	for _, card := range cards {
		if card.Priority < priority {
			exist = false
			break
		}
		if card.Priority == priority {
			exist = true
			break
		}
	}
	return exist
}

//获取房间匹配信息
func (r *Room) getRoomMatchingInfo() ([]string, []string) {
	//已落座和未落座的玩家集合
	userids := r.getAllUserIDs()
	//获取已落座的玩家状态
	getUsersStatuss := func() string {
		bfStatuss := bytes.Buffer{}
		for _, user := range r.users {
			if user != nil {
				userInfo, err := redis.Ints(user.GetUserInfo("vip", "level"))
				logger.CheckError(err)
				vip, level := userInfo[0], userInfo[1]
				bfStatuss.WriteString(fmt.Sprintf("%s$%d$%d$%d$%d|", *user.userid, user.getStatus(), vip, level, user.getHYTWIntegral()))
			} else {
				bfStatuss.WriteString("|")
			}
		}
		strStatuss := bfStatuss.String()
		strStatuss = strStatuss[:len(strStatuss)-1]
		return strStatuss
	}
	//获取未落座的玩家状态
	getIdleUsersStatuss := func() string {
		bfStatuss := bytes.Buffer{}
		for _, user := range r.idleusers {
			bfStatuss.WriteString(fmt.Sprintf("%s|", *user.userid))
		}
		strStatuss := bfStatuss.String()
		if strStatuss != "" {
			strStatuss = strStatuss[:len(strStatuss)-1]
		}
		return strStatuss
	}
	statuss := []string{fmt.Sprintf("%s#%s#%d", getUsersStatuss(), getIdleUsersStatuss(), r.getInning())}
	return userids, statuss
}

//设置所有人都过牌
func (r *Room) setAllUserPass() {
	for _, user := range r.users {
		user.setStatus(UserStatus_Pass)
	}
}

//设置所有人都没过牌
func (r *Room) setAllUserNoPass() {
	for _, user := range r.users {
		user.setStatus(UserStatus_NoPass)
	}
}

//开启新的一轮
func (r *Room) newCycle() {
	r.setCurrentCards([]Card{})
	r.setCurrentCardsUser(nil)
	for _, user := range r.users {
		user.setStatus(UserStatus_NoPass)
	}
	//某一轮的出牌信息
	r.users_cards = make(map[string]string, pcount)
}

//记录玩家的出牌信息(只保存一轮),为了重新进游戏
func (r *Room) RecordUserPlayCardInfo(user *User, cardInfo *string) {
	r.users_cards[*user.getUserID()] = *cardInfo
}
