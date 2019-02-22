package engine

import (
	"fmt"
	"reflect"
)

//检测任务是否完成
func (u *User) checkTaskFinish() {
	//已完成,不再检测了
	if u.getTaskFinish() {
		return
	}
	room := u.getRoom()
	task := room.getCouncilTask()
	v := reflect.ValueOf(u)
	funcName := fmt.Sprintf("CheckTask%d", task.getID())
	mv := v.MethodByName(funcName)
	mv.Call(nil)
}

//自己打出两幅三带一并获得胜利
func (u *User) CheckTask0() {
	count := u.getCardTypeCount(CARDTYPE_SANDAIYI)
	if count >= 2 {
		u.setTaskFinish(true)
	}
}

//自己打出两幅三带二并获得胜利
func (u *User) CheckTask1() {
	count := u.getCardTypeCount(CARDTYPE_SANDAIER)
	if count >= 2 {
		u.setTaskFinish(true)
	}
}

//打出一副炸弹并获得胜利
func (u *User) CheckTask2() {
	zhadanCount := u.getCardTypeCount(CARDTYPE_ZHADAN)
	huojianCount := u.getCardTypeCount(CARDTYPE_HUOJIAN)
	if zhadanCount+huojianCount >= 1 {
		u.setTaskFinish(true)
	}
}

//最后一手牌打出飞机并获得胜利
func (u *User) CheckTask3() {
	if len(u.getCards()) <= 0 {
		currentCards := u.getRoom().getCurrentCards()
		if cardType, _ := u.getCardType(currentCards); cardType == CARDTYPE_FEIJI {
			u.setTaskFinish(true)
		}
	}
}

//最后一手牌打出小王并获得胜利
func (u *User) CheckTask4() {
	if len(u.getCards()) <= 0 {
		currentCards := u.getRoom().getCurrentCards()
		if len(currentCards) == 1 {
			if currentCards[0].Priority == Priority_SKing {
				u.setTaskFinish(true)
			}
		}
	}
}

//自己打出一副飞机并获得胜利
func (u *User) CheckTask5() {
	count := u.getCardTypeCount(CARDTYPE_FEIJI)
	if count >= 1 {
		u.setTaskFinish(true)
	}
}

//第一手牌打出一副顺子并获得胜利
func (u *User) CheckTask6() {
	if u.getPlayCardCount() == 0 {
		currentCards := u.getRoom().getCurrentCards()
		if cardType, _ := u.getCardType(currentCards); cardType == CARDTYPE_SHUNZI {
			u.setTaskFinish(true)
		}
	}
}

//自己打出两幅顺子并获得胜利
func (u *User) CheckTask7() {
	count := u.getCardTypeCount(CARDTYPE_SHUNZI)
	if count >= 2 {
		u.setTaskFinish(true)
	}
}

//最后一手牌打出梅花Q并获得胜利
func (u *User) CheckTask8() {
	if len(u.getCards()) <= 0 {
		currentCards := u.getRoom().getCurrentCards()
		if len(currentCards) == 1 {
			card := currentCards[0]
			if card.Priority == Priority_Twelve && card.Suit == SUIT_MEIHUA {
				u.setTaskFinish(true)
			}
		}
	}
}

//最后一手牌打出大王并获得胜利
func (u *User) CheckTask9() {
	if len(u.getCards()) <= 0 {
		currentCards := u.getRoom().getCurrentCards()
		if len(currentCards) == 1 {
			if currentCards[0].Priority == Priority_BKing {
				u.setTaskFinish(true)
			}
		}
	}
}

//最后一手牌打出一副顺子并获得胜利
func (u *User) CheckTask10() {
	if len(u.getCards()) <= 0 {
		currentCards := u.getRoom().getCurrentCards()
		if cardType, _ := u.getCardType(currentCards); cardType == CARDTYPE_SHUNZI {
			u.setTaskFinish(true)
		}
	}
}

//自己打出一副王炸并获得胜利
func (u *User) CheckTask11() {
	count := u.getCardTypeCount(CARDTYPE_HUOJIAN)
	if count >= 1 {
		u.setTaskFinish(true)
	}
}
