/*
玩家-操作-出牌
*/

package engine

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"

	. "kelei.com/utils/common"
	"kelei.com/utils/logger"
)

/*
玩家出牌
in:牌index列表
out:-1没有选择牌 -2你的手中没有这套牌 -3牌值无效 -5牌太小 -103玩家没有操作权限
	1成功
*/
func PlayCard(args []string) *string {
	fmt.Print("")
	res := Res_Succeed
	userid := args[0]
	user := UserManage.GetUser(&userid)
	//玩家没有操作权限
	if !user.getHandlePerm() {
		return &Res_NoPerm
	}
	play_indexs_str := args[1]
	if play_indexs_str == "" {
		res = "-1"
		return &res
	}
	//将字符串的index数组转化成数字index数组
	play_indexs := StrArrToIntArr(strings.Split(play_indexs_str, "|"))
	sort.Ints(play_indexs)
	if user.checkIndex(play_indexs) {
		res = "-2"
		return &res
	}
	userCards := user.getCards()
	play_cards := indexsToCards(play_indexs, userCards)
	cardType, _ := user.getCardType(play_cards)
	//牌值无效
	if cardType == -1 {
		res = "-3"
		return &res
	}
	room := user.getRoom()
	current_cards := room.getCurrentCards()
	//压牌
	if len(current_cards) > 0 {
		if !user.excel(current_cards, play_cards) {
			res = "-5"
			return &res
		}
	}
	//关闭玩家操作倒计时
	user.close_countDown_playCard()
	//更新玩家手中的牌
	user.updateUserCards(play_indexs)
	//将牌打出
	user.play(play_cards, cardType)
	//翻倍
	user.DoubleByCardType(cardType)
	//记录(地主、农民)出牌次数
	user.recordPlayCardCount()
	//检测比赛是否结束
	room.checkMatchingOver()
	//将牌打出之后,后续的操作
	user.playAfter(play_cards)
	return &res
}

//记录(地主、农民)出牌次数
func (u *User) recordPlayCardCount() {
	room := u.getRoom()
	if u.getLandlord() {
		room.updteLandlordPlayCardCount()
	} else {
		room.updteFarmerPlayCardCount()
	}
}

//翻倍
func (u *User) DoubleByCardType(cardType int) {
	if cardType == CARDTYPE_HUOJIAN || cardType == CARDTYPE_ZHADAN {
		room := u.getRoom()
		room.doubleMultiple()
	}
}

/*
打出
push:Play,userid,cardid|cardid,出牌类型,牌类型
des:出牌类型(0正常出牌 1残局进来的出牌)
*/
func (u *User) play(playCards []Card, cardType int) {
	room := u.getRoom()
	userids := room.getUserIDs()
	cardids := *getCardIDs(playCards)
	room.setCurrentCards(playCards)
	room.setCurrentCardsUser(u)
	//记录玩家的出牌信息(只保存一轮),为了重新进游戏
	room.RecordUserPlayCardInfo(u, &cardids)
	//打出的牌型记录
	u.addCardTypeCount(cardType)
	//判定任务是否完成
	u.checkTaskFinish()
	//更新出牌次数
	u.updatePlayCardCount()
	//记牌器更新
	u.rememberCard(playCards)
	//记录出牌信息
	u.recordPlayCard(playCards)
	//向所有人推送出牌的信息
	message := fmt.Sprintf("%s,%s,%d,%d", *u.getUserID(), cardids, PlayType_Normal, cardType)
	pushMessageToUsers("Play", []string{message}, userids)
	room.pushJudgment("Play", message)
}

//记牌器更新
func (u *User) rememberCard(playCards []Card) {
	room := u.getRoom()
	for _, card := range playCards {
		if card.Priority < Priority_Two {
			break
		}
		if card.Priority == Priority_Two {
			room.updateSurplusTwoCount()
		} else if card.Priority == Priority_SKing {
			room.updateSurplusSKingCount()
		} else if card.Priority == Priority_BKing {
			room.updateSurplusBKingCount()
		}
	}
}

//记录出牌信息
func (u *User) recordPlayCard(play_cards []Card) {
	room := u.getRoom()
	//记录出牌的信息（debug）
	playTime := room.updatePlayTime()
	bf := bytes.Buffer{}
	for _, card := range play_cards {
		bf.WriteString(strconv.Itoa(card.Value))
		bf.WriteString("|")
	}
	bf1 := bytes.Buffer{}
	for _, card := range u.getCards() {
		bf1.WriteString(strconv.Itoa(card.Value))
		bf1.WriteString("|")
	}
	//记录牌
	logger.Debugf("第（%d）次出牌:%s,剩余牌:%s", playTime, bf.String(), bf1.String())
}

//打出之后
func (u *User) playAfter(playCards []Card) {
	room := u.getRoom()
	if room == nil {
		return
	}
	//比赛结束
	if !u.isMatching() {
		return
	}
	//如果打出的牌无敌
	if u.cardInvincible(playCards) {
		room.currentCardsUserNewCycle()
		return
	}
	//设置所有人未过牌
	room.setAllUserNoPass()
	//获取下一顺位的玩家
	nextUser := room.getNextUser()
	room.setController(nextUser, SetController_Press)
}

//牌是否无敌
func (u *User) cardInvincible(playCards []Card) bool {
	cardType, _ := u.getCardType(playCards)
	if cardType == CARDTYPE_HUOJIAN {
		return true
	}
	return false
}

//将出的牌的index列表转化成对应的牌列表
func indexsToCards(indexs []int, userCards []Card) []Card {
	cards := make([]Card, len(indexs))
	for i, index := range indexs {
		cards[i] = userCards[index]
	}
	return cards
}

//选择的牌是否索引越界
func (u *User) checkIndex(indexs []int) bool {
	return indexs[len(indexs)-1]+1 > len(u.cards)
}

func (u *User) Excel() {
	a := []Card{Card{Priority: 6}, Card{Priority: 6}, Card{Priority: 6}, Card{Priority: 6}, Card{Priority: 5}, Card{Priority: 5}}
	b := []Card{Card{Priority: 4}, Card{Priority: 4}, Card{Priority: 4}, Card{Priority: 4}, Card{Priority: 2}, Card{Priority: 2}}
	fmt.Println("xxxxx:", u.excel(a, b))
}

//检测牌是否能压过(当前的牌,要出的牌)
func (u *User) excel(a, b []Card) bool {
	aType, aCardTypeStruct := u.getCardType(a)
	bType, bCardTypeStruct := u.getCardType(b)
	if aType == CARDTYPE_HUOJIAN {
		return false
	} else if bType == CARDTYPE_HUOJIAN {
		return true
	} else if aType == CARDTYPE_ZHADAN || bType == CARDTYPE_ZHADAN {
		if aType == CARDTYPE_ZHADAN && bType == CARDTYPE_ZHADAN {
			if bCardTypeStruct.CruxCard.Priority > aCardTypeStruct.CruxCard.Priority {
				return true
			}
		} else if aType == CARDTYPE_ZHADAN {
			return false
		} else if bType == CARDTYPE_ZHADAN {
			return true
		}
	} else if aType == bType {
		if bCardTypeStruct.CruxCard.Priority > aCardTypeStruct.CruxCard.Priority {
			return true
		}
	}
	return false
}
