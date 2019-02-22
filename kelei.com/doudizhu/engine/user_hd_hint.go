/*
玩家-操作-提示
*/

package engine

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"

	"kelei.com/utils/logger"
)

/*
Hint(提示)
result:
	有能压过的牌:index|index|index
	没有:-2
*/
func Hint(args []string) *string {
	fmt.Print("")
	res := "-2" //压不过
	userid := args[0]
	user := UserManage.GetUser(&userid)
	room := user.getRoom()
	current_cards := room.getCurrentCards()
	playCards := []Card{}
	//不是压牌
	if len(current_cards) == 0 {
		playCards = user.getLeastCards()
	} else {
		//是压牌
		playCards = user.getPressIndexs()
	}
	if len(playCards) > 0 {
		bf := bytes.Buffer{}
		for _, card := range playCards {
			bf.WriteString(strconv.Itoa(card.Index) + "|")
		}
		str := bf.String()
		res = str[:len(str)-1]
	}
	return &res
}

func (u *User) GetLeastCards() {
}

//获取最小牌
func (u *User) getLeastCards() []Card {
	lintCards := []Card{}
	//最小炸弹
	minZhadan := []Card{}
	//玩家牌面
	cards := u.getCards()
	for i := len(cards) - 1; i >= 0; i-- {
		card := cards[i]
		if len(lintCards) > 0 && card.Priority != lintCards[0].Priority {
			//是炸弹,获取下一套
			if u.IsZhaDan(lintCards) {
				//最小炸弹赋值
				if len(minZhadan) == 0 {
					minZhadan = lintCards
				}
				lintCards = []Card{}
			} else {
				break
			}
		}
		lintCards = append(lintCards, card)
	}
	//最小的牌还是炸弹,就是用最小的炸弹
	if u.IsZhaDan(lintCards) {
		if len(minZhadan) > 0 {
			lintCards = minZhadan
		}
	}
	return lintCards
}

func (u *User) GetPressIndexs() {
	room := Room{}
	u.setRoom(&room)
	//火箭
	room.setCurrentCards([]Card{Card{Priority: 15}, Card{Priority: 14}})
	u.setCards([]Card{Card{Priority: 13}, Card{Priority: 13}, Card{Priority: 13}, Card{Priority: 13}})
	fmt.Println("压牌:", u.getPressIndexs())
	//炸弹
	room.setCurrentCards([]Card{Card{Priority: 12}, Card{Priority: 12}, Card{Priority: 12}, Card{Priority: 12}})
	u.setCards([]Card{Card{Priority: 15}, Card{Priority: 14}, Card{Priority: 13}, Card{Priority: 13}, Card{Priority: 13}, Card{Priority: 13}})
	fmt.Println("压牌:", u.getPressIndexs())
	//单
	room.setCurrentCards([]Card{Card{Priority: 15}})
	u.setCards([]Card{Card{Priority: 11}, Card{Priority: 11}, Card{Priority: 11}, Card{Priority: 11}})
	fmt.Println("压牌:", u.getPressIndexs())
	//对
	room.setCurrentCards([]Card{Card{Priority: 10}, Card{Priority: 10}})
	u.setCards([]Card{Card{Priority: 1}, Card{Priority: 1}, Card{Priority: 1}, Card{Priority: 1}})
	fmt.Println("压牌:", u.getPressIndexs())
	//三
	room.setCurrentCards([]Card{Card{Priority: 10}, Card{Priority: 10}, Card{Priority: 10}})
	u.setCards([]Card{Card{Priority: 15}, Card{Priority: 14}, Card{Priority: 11}, Card{Priority: 11}, Card{Priority: 11}, Card{Priority: 10}})
	fmt.Println("压牌:", u.getPressIndexs())
	//三带一
	room.setCurrentCards([]Card{Card{Priority: 15}, Card{Priority: 10}, Card{Priority: 10}, Card{Priority: 10}})
	u.setCards([]Card{Card{Priority: 15}, Card{Priority: 14}, Card{Priority: 11}, Card{Priority: 11}, Card{Priority: 11}})
	fmt.Println("压牌:", u.getPressIndexs())
	//三带二
	room.setCurrentCards([]Card{Card{Priority: 10}, Card{Priority: 10}, Card{Priority: 10}, Card{Priority: 1}, Card{Priority: 1}})
	u.setCards([]Card{Card{Priority: 15}, Card{Priority: 14}, Card{Priority: 11}, Card{Priority: 11}, Card{Priority: 11}, Card{Priority: 10}, Card{Priority: 9}})
	fmt.Println("压牌:", u.getPressIndexs())
	//顺子
	room.setCurrentCards([]Card{Card{Priority: 6}, Card{Priority: 5}, Card{Priority: 4}, Card{Priority: 3}, Card{Priority: 2}})
	u.setCards([]Card{Card{Priority: 8}, Card{Priority: 7}, Card{Priority: 6}, Card{Priority: 5}, Card{Priority: 4}, Card{Priority: 4}, Card{Priority: 4}, Card{Priority: 3}, Card{Priority: 2}})
	fmt.Println("压牌:", u.getPressIndexs())
	//连对
	room.setCurrentCards([]Card{Card{Priority: 3}, Card{Priority: 3}, Card{Priority: 2}, Card{Priority: 2}, Card{Priority: 1}, Card{Priority: 1}})
	u.setCards([]Card{Card{Priority: 12}, Card{Priority: 11}, Card{Priority: 10}, Card{Priority: 10}, Card{Priority: 9}, Card{Priority: 9}, Card{Priority: 8}, Card{Priority: 6}, Card{Priority: 6}, Card{Priority: 5}, Card{Priority: 5}, Card{Priority: 4}, Card{Priority: 4}, Card{Priority: 2}, Card{Priority: 2}, Card{Priority: 1}, Card{Priority: 1}})
	fmt.Println("压牌:", u.getPressIndexs())
	//飞机
	room.setCurrentCards([]Card{Card{Priority: 1}, Card{Priority: 2}, Card{Priority: 3}, Card{Priority: 6}, Card{Priority: 6}, Card{Priority: 6}, Card{Priority: 5}, Card{Priority: 5}, Card{Priority: 5}, Card{Priority: 4}, Card{Priority: 4}, Card{Priority: 4}})
	u.setCards([]Card{Card{Priority: 9}, Card{Priority: 9}, Card{Priority: 9}, Card{Priority: 8}, Card{Priority: 8}, Card{Priority: 8}, Card{Priority: 7}, Card{Priority: 7}, Card{Priority: 7}, Card{Priority: 6}, Card{Priority: 11}, Card{Priority: 6}})
	fmt.Println("压牌:", u.getPressIndexs())
	//四带二
	room.setCurrentCards([]Card{Card{Priority: 6}, Card{Priority: 6}, Card{Priority: 6}, Card{Priority: 6}, Card{Priority: 5}, Card{Priority: 5}})
	u.setCards([]Card{Card{Priority: 7}, Card{Priority: 7}, Card{Priority: 7}, Card{Priority: 7}, Card{Priority: 2}, Card{Priority: 2}})
	fmt.Println("压牌:", u.getPressIndexs())
	//	common.CostTime(func() {
	//		u.getPressIndexs()
	//	}, 10000, "---")
}

//获取压牌的信息
func (u *User) getPressIndexs() []Card {
	lintCards := []Card{}
	currentCards := u.getRoom().getCurrentCards()
	cardType, cardTypeStruct := u.getCardType(currentCards)
	if cardType == -1 {
		logger.Debugf("---------牌类型:无")
	} else {
		logger.Debugf("---------牌类型:%s", CardTypeNames[cardType])
	}
	switch cardType {
	case CARDTYPE_DAN:
		lintCards = u.dan()
	case CARDTYPE_DUI:
		lintCards = u.dui()
	case CARDTYPE_SANDAIYI:
		lintCards = u.sandaiyi()
	case CARDTYPE_SANDAIER:
		lintCards = u.sandaier()
	case CARDTYPE_SHUNZI:
		lintCards = u.shunzi()
	case CARDTYPE_LIANDUI:
		lintCards = u.liandui()
	case CARDTYPE_FEIJI:
		lintCards = u.feiji()
	case CARDTYPE_SAN:
		lintCards = u.san()
	case CARDTYPE_ZHADAN:
		lintCards = u.zhadan()
	case CARDTYPE_HUOJIAN:
		lintCards = u.huojian()
	case CARDTYPE_SIDAIER:
		lintCards = u.sidaier(cardTypeStruct)
	}
	//当前牌不是火箭
	if cardType != CARDTYPE_HUOJIAN {
		if cardType == CARDTYPE_ZHADAN { //当前牌是炸弹
			//有火箭
			if u.haveHuojian() {
				lintCards = u.getHuojian()
			}
		} else { //当前牌不是炸弹
			//没有能压过的同类型牌
			if len(lintCards) == 0 {
				//找一个最小的炸弹
				lintCards = u.getFour(nil)
				if len(lintCards) == 0 {
					lintCards = u.getHuojian()
				}
			}
		}
	}
	return lintCards
}

//筛选
type Filter struct {
	BaseCard     *Card    //大于此牌的牌
	Count        int      //0获取所有数量的 >0大于指定数量的牌
	Limit        int      //0所有牌 1一套牌
	ExcludeCards [][]Card //排除的牌
	FromCards    []Card   //从这些牌中筛选,nil从玩家手牌中筛选
}

//筛选结果
type FilterResult struct {
	ones   [][]Card //一张牌的列表
	twos   [][]Card //两张牌的列表
	threes [][]Card //三张牌的列表
}

//获取火箭
func (u *User) getHuojian() []Card {
	lintCards := []Card{}
	userCards := u.getCards()
	if userCards[0].Priority == Priority_BKing && userCards[1].Priority == Priority_SKing {
		lintCards = userCards[:2]
	}
	return lintCards
}

//是否有火箭
func (u *User) haveHuojian() bool {
	return len(u.getHuojian()) > 0
}

//获取炸弹
func (u *User) getFour(baseCard *Card) []Card {
	tmpCards := []Card{}
	userCards := u.getCards()
	for i := len(userCards) - 1; i >= 0; i-- {
		userCard := userCards[i]
		//		if baseCard!=nil &&
		if len(tmpCards) > 0 && userCard.Priority != tmpCards[0].Priority {
			if len(tmpCards) >= 4 {
				return tmpCards
			}
			tmpCards = []Card{}
		}
		if baseCard == nil {
			tmpCards = append(tmpCards, userCard)
		} else {
			if userCard.Priority > baseCard.Priority {
				tmpCards = append(tmpCards, userCard)
			}
		}
	}
	if len(tmpCards) >= 4 {
		return tmpCards
	}
	return []Card{}
}

//获取1、2、3张牌的信息
func (u *User) getOneTwoThree(filter *Filter) *FilterResult {
	filterResult := &FilterResult{}
	tmpCards := []Card{}
	userCards := []Card{}
	if filter.FromCards == nil {
		userCards = u.getCards()
	} else {
		userCards = filter.FromCards
	}
	handle := func() {
		appendCard := func() {
			if len(tmpCards) == 1 {
				if filter.Limit == 1 && len(filterResult.ones) > 0 {
					return
				}
				filterResult.ones = append(filterResult.ones, tmpCards)
			} else if len(tmpCards) == 2 {
				if filter.Limit == 1 && len(filterResult.twos) > 0 {
					return
				}
				filterResult.twos = append(filterResult.twos, tmpCards)
			} else if len(tmpCards) == 3 {
				if filter.Limit == 1 && len(filterResult.threes) > 0 {
					return
				}
				filterResult.threes = append(filterResult.threes, tmpCards)
			}
		}
		if filter == nil {
			appendCard()
		} else {
			if filter.BaseCard == nil && filter.Count == 0 {
				appendCard()
			} else if filter.BaseCard != nil && filter.Count == 0 {
				if tmpCards[0].Priority > filter.BaseCard.Priority {
					appendCard()
				}
			} else if filter.BaseCard == nil && filter.Count != 0 {
				if len(tmpCards) >= filter.Count {
					appendCard()
				}
			} else if filter.BaseCard != nil && filter.Count != 0 {
				if tmpCards[0].Priority > filter.BaseCard.Priority && len(tmpCards) >= filter.Count {
					appendCard()
				}
			}
		}
	}
	for i := len(userCards) - 1; i >= 0; i-- {
		userCard := userCards[i]
		if len(filter.ExcludeCards) > 0 {
			exclude := false
			for _, excludeCard := range filter.ExcludeCards {
				if excludeCard[0].Priority == userCard.Priority {
					exclude = true
					break
				}
			}
			if exclude {
				continue
			}
		}
		if len(tmpCards) > 0 && userCard.Priority != tmpCards[0].Priority {
			handle()
			tmpCards = []Card{}
		}
		tmpCards = append(tmpCards, userCard)
	}
	handle()
	return filterResult
}

//单处理
func (u *User) dan() []Card {
	lintCards := []Card{}
	currentCards := u.getRoom().getCurrentCards()
	fitler := &Filter{BaseCard: &currentCards[0], Limit: 1}
	filterResult := u.getOneTwoThree(fitler)
	if len(filterResult.ones) > 0 {
		lintCards = filterResult.ones[0]
	} else if len(filterResult.twos) > 0 {
		lintCards = filterResult.twos[0][:1]
	} else if len(filterResult.threes) > 0 {
		lintCards = filterResult.threes[0][:1]
	}
	return lintCards
}

//对处理
func (u *User) dui() []Card {
	lintCards := []Card{}
	currentCards := u.getRoom().getCurrentCards()
	fitler := &Filter{BaseCard: &currentCards[0], Count: 2, Limit: 1}
	filterResult := u.getOneTwoThree(fitler)
	if len(filterResult.twos) > 0 {
		lintCards = filterResult.twos[0][:2]
	} else if len(filterResult.threes) > 0 {
		lintCards = filterResult.threes[0][:2]
	}
	return lintCards
}

//三带一处理
func (u *User) sandaiyi() []Card {
	lintCards := []Card{}
	currentCards := u.getRoom().getCurrentCards()
	_, currentCardTypeStruct := u.getCardType(currentCards)
	fitler := &Filter{BaseCard: currentCardTypeStruct.CruxCard, Count: 3, Limit: 1}
	filterResult := u.getOneTwoThree(fitler)
	if len(filterResult.threes) == 0 {
		return lintCards
	}
	lintCards = append(lintCards, filterResult.threes[0]...)
	fitler = &Filter{ExcludeCards: filterResult.threes}
	filterResult = u.getOneTwoThree(fitler)
	if len(filterResult.ones) > 0 {
		lintCards = append(lintCards, filterResult.ones[0]...)
	} else if len(filterResult.twos) > 0 {
		lintCards = append(lintCards, filterResult.twos[0][:1]...)
	} else if len(filterResult.threes) > 0 {
		lintCards = append(lintCards, filterResult.threes[0][:1]...)
	}
	if len(lintCards) != 4 {
		lintCards = []Card{}
	}
	return lintCards
}

//三带二处理
func (u *User) sandaier() []Card {
	lintCards := []Card{}
	currentCards := u.getRoom().getCurrentCards()
	_, currentCardTypeStruct := u.getCardType(currentCards)
	fitler := &Filter{BaseCard: currentCardTypeStruct.CruxCard, Count: 3, Limit: 1}
	filterResult := u.getOneTwoThree(fitler)
	if len(filterResult.threes) == 0 {
		return lintCards
	}
	lintCards = append(lintCards, filterResult.threes[0]...)
	fitler = &Filter{ExcludeCards: filterResult.threes}
	filterResult = u.getOneTwoThree(fitler)
	if len(filterResult.twos) > 0 {
		lintCards = append(lintCards, filterResult.twos[0][:2]...)
	} else if len(filterResult.threes) > 0 {
		lintCards = append(lintCards, filterResult.threes[0][:2]...)
	}
	if len(lintCards) != 5 {
		lintCards = []Card{}
	}
	return lintCards
}

//顺子处理
func (u *User) shunzi() []Card {
	lintCards := []Card{}
	currentCards := u.getRoom().getCurrentCards()
	minCard := currentCards[len(currentCards)-1]
	userCards := u.getCards()
	tmpCards := []Card{}
	handle := func() {
		if len(lintCards) >= len(currentCards) {
			return
		}
		if len(tmpCards) < 4 && tmpCards[0].Priority > minCard.Priority {
			if tmpCards[0].Priority < Priority_Two {
				if len(lintCards) == 0 {
					lintCards = append(lintCards, tmpCards[0])
				} else {
					//是顺序的
					if tmpCards[0].Priority == lintCards[len(lintCards)-1].Priority+1 {
						lintCards = append(lintCards, tmpCards[0])
					} else {
						lintCards = []Card{}
						lintCards = append(lintCards, tmpCards[0])
					}
				}
			}
		}
	}
	for i := len(userCards) - 1; i >= 0; i-- {
		userCard := userCards[i]
		if len(tmpCards) > 0 && userCard.Priority != tmpCards[0].Priority {
			handle()
			tmpCards = []Card{}
		}
		tmpCards = append(tmpCards, userCard)
	}
	handle()
	if len(lintCards) < len(currentCards) {
		lintCards = []Card{}
	}
	return lintCards
}

//连对处理
func (u *User) liandui() []Card {
	lintCards := []Card{}
	currentCards := u.getRoom().getCurrentCards()
	minCard := currentCards[len(currentCards)-1]
	userCards := u.getCards()
	tmpCards := []Card{}
	handle := func() {
		if len(lintCards) >= len(currentCards) {
			return
		}
		if len(tmpCards) >= 2 && len(tmpCards) < 4 && tmpCards[0].Priority > minCard.Priority {
			if tmpCards[0].Priority < Priority_Two {
				if len(lintCards) == 0 {
					lintCards = append(lintCards, tmpCards[0], tmpCards[1])
				} else {
					//是顺序的
					if tmpCards[0].Priority == lintCards[len(lintCards)-1].Priority+1 {
						lintCards = append(lintCards, tmpCards[0], tmpCards[1])
					} else {
						lintCards = []Card{}
						lintCards = append(lintCards, tmpCards[0], tmpCards[1])
					}
				}
			}
		}
	}
	for i := len(userCards) - 1; i >= 0; i-- {
		userCard := userCards[i]
		if len(tmpCards) > 0 && userCard.Priority != tmpCards[0].Priority {
			handle()
			tmpCards = []Card{}
		}
		tmpCards = append(tmpCards, userCard)
	}
	handle()
	if len(lintCards) < len(currentCards) {
		lintCards = []Card{}
	}
	return lintCards
}

//飞机处理
func (u *User) feiji() []Card {
	lintCards := []Card{}
	currentCards := u.getRoom().getCurrentCards()
	minCard := currentCards[len(currentCards)-1]
	userCards := u.getCards()
	tmpCards := []Card{}
	filter := Filter{FromCards: currentCards}
	filterResult := u.getOneTwoThree(&filter)
	threesCards := [][]Card{}
	func() {
		priority := 0
		feijiCount := 0
		for _, three := range filterResult.threes {
			cardPriority := three[0].Priority
			if priority == 0 {
				priority = cardPriority
				feijiCount += 1
				threesCards = append(threesCards, three)
			} else {
				if priority+1 == cardPriority {
					priority = cardPriority
					feijiCount += 1
					threesCards = append(threesCards, three)
				} else {
					if feijiCount <= 1 {
						priority = cardPriority
						feijiCount = 1
						threesCards = [][]Card{}
						threesCards = append(threesCards, three)
					} else {
						break
					}
				}
			}
		}
	}()
	filter = Filter{FromCards: currentCards, ExcludeCards: threesCards}
	otherThreesCards := u.getOneTwoThree(&filter).threes
	handle := func() {
		if len(lintCards)/3 >= len(threesCards) {
			return
		}
		if len(tmpCards) == 3 && tmpCards[0].Priority > minCard.Priority {
			if tmpCards[0].Priority < Priority_Two {
				if len(lintCards) == 0 {
					lintCards = append(lintCards, tmpCards[0], tmpCards[1], tmpCards[2])
				} else {
					//是顺序的
					if tmpCards[0].Priority == lintCards[len(lintCards)-1].Priority+1 {
						lintCards = append(lintCards, tmpCards[0], tmpCards[1], tmpCards[2])
					} else {
						lintCards = []Card{}
						lintCards = append(lintCards, tmpCards[0], tmpCards[1], tmpCards[2])
					}
				}
			}
		}
	}
	for i := len(userCards) - 1; i >= 0; i-- {
		userCard := userCards[i]
		if len(tmpCards) > 0 && userCard.Priority != tmpCards[0].Priority {
			handle()
			tmpCards = []Card{}
		}
		tmpCards = append(tmpCards, userCard)
	}
	handle()
	//飞机是否够长
	if len(lintCards)/3 >= len(threesCards) {
		//截取飞机
		lintCards = lintCards[:len(threesCards)*3]
		//将飞机转化成[][]Card{}格式
		threes := [][]Card{}
		for i, lintCard := range lintCards {
			if (i+1)%3 == 1 {
				threes = append(threes, []Card{lintCard})
			}
		}
		//飞机数量
		feijiCount := len(threes)
		//获取除去飞机的其他牌
		myFilter := Filter{ExcludeCards: threes}
		userFilterResult := u.getOneTwoThree(&myFilter)
		//当前牌的翅膀是单牌 || 当前牌的翅膀是单牌（两张一样的牌）
		if len(filterResult.ones) > 0 || (len(filterResult.twos) > 0 && feijiCount/len(filterResult.twos) == 2) {
			oneCards := u.convertCardToX(userFilterResult, 1)
			if len(oneCards) >= feijiCount {
				for i := 0; i < feijiCount; i++ {
					lintCards = append(lintCards, oneCards[i][0])
				}
			} else {
				lintCards = []Card{}
			}
		} else if len(filterResult.twos) > 0 {
			twoCards := u.convertCardToX(userFilterResult, 2)
			if len(twoCards) >= feijiCount {
				for i := 0; i < feijiCount; i++ {
					lintCards = append(lintCards, twoCards[i][:2]...)
				}
			} else {
				lintCards = []Card{}
			}
		} else if len(otherThreesCards) > 0 {
			oneCards := u.convertCardToX(userFilterResult, 1)
			if len(oneCards) >= feijiCount {
				for i := 0; i < feijiCount; i++ {
					lintCards = append(lintCards, oneCards[i][0])
				}
			} else {
				lintCards = []Card{}
			}
		}
	} else {
		lintCards = []Card{}
	}
	return lintCards
}

//将1.2.3张牌都转化成x张牌
func (u *User) convertCardToX(filterResult *FilterResult, x int) [][]Card {
	result := [][]Card{}
	if x == 1 {
		for _, one := range filterResult.ones {
			if one[0].Priority <= Priority_Two {
				result = append(result, one)
			} else {
				if !u.haveHuojian() {
					result = append(result, one)
				}
			}
		}
		for _, two := range filterResult.twos {
			result = append(result, two[:1], two[1:2])
		}
		for _, three := range filterResult.threes {
			result = append(result, three[:1], three[1:2], three[2:3])
		}
	} else if x == 2 {
		for _, two := range filterResult.twos {
			result = append(result, two)
		}
		for _, three := range filterResult.threes {
			result = append(result, three[:2])
		}
	} else if x == 3 {
		result = filterResult.threes
	}
	return result
}

//三处理
func (u *User) san() []Card {
	lintCards := []Card{}
	currentCards := u.getRoom().getCurrentCards()
	fitler := &Filter{BaseCard: &currentCards[0], Count: 3, Limit: 1}
	filterResult := u.getOneTwoThree(fitler)
	if len(filterResult.threes) > 0 {
		lintCards = filterResult.threes[0][:3]
	}
	return lintCards
}

//炸弹处理
func (u *User) zhadan() []Card {
	lintCards := []Card{}
	currentCards := u.getRoom().getCurrentCards()
	minCard := currentCards[len(currentCards)-1]
	userCards := u.getCards()
	tmpCards := []Card{}
	handle := func() {
		if len(tmpCards) >= 4 && tmpCards[0].Priority > minCard.Priority {
			if len(lintCards) == 0 {
				lintCards = tmpCards
			}
		}
	}
	for i := len(userCards) - 1; i >= 0; i-- {
		userCard := userCards[i]
		if len(tmpCards) > 0 && userCard.Priority != tmpCards[0].Priority {
			handle()
			tmpCards = []Card{}
		}
		tmpCards = append(tmpCards, userCard)
	}
	handle()
	return lintCards
}

//火箭处理
func (u *User) huojian() []Card {
	lintCards := []Card{}
	return lintCards
}

//四带二
func (u *User) sidaier(cardTypeStruct *CardTypeStruct) []Card {
	lintCards := []Card{}
	four := u.getFour(cardTypeStruct.CruxCard)
	if len(four) > 0 {
		lintCards = four
		currentCards := u.getRoom().getCurrentCards()
		filter := &Filter{FromCards: currentCards}
		filterResult := u.getOneTwoThree(filter)
		myFilter := Filter{ExcludeCards: [][]Card{four}}
		myFilterResult := u.getOneTwoThree(&myFilter)
		//两张单
		if len(filterResult.ones) > 0 || len(filterResult.twos) == 1 {
			oneCards := u.convertCardToX(myFilterResult, 1)
			if len(oneCards) >= 2 {
				//				if oneCards[0][0].Priority > cardTypeStruct.CruxCard.Priority {
				for i := 0; i < 2; i++ {
					lintCards = append(lintCards, oneCards[i][0])
				}
				//				}
			} else {
				lintCards = []Card{}
			}
		} else if len(filterResult.twos) > 0 { //两个对
			twoCards := u.convertCardToX(myFilterResult, 2)
			if len(twoCards) >= 2 {
				//				if oneCards[0][0].Priority > cardTypeStruct.CruxCard.Priority {
				for i := 0; i < 2; i++ {
					lintCards = append(lintCards, twoCards[i][:2]...)
				}
				//				}
			} else {
				lintCards = []Card{}
			}
		}
	}
	return lintCards
}

//牌结构
type CardTypeStruct struct {
	CruxCard *Card //关键牌,用来比较两套牌的大小
}

//获取牌是什么类型
func (u *User) getCardType(cards []Card) (int, *CardTypeStruct) {
	cardType := -1
	var cardTypeStruct *CardTypeStruct
	if u.IsDan(cards) {
		cardType = CARDTYPE_DAN
		cardTypeStruct = &CardTypeStruct{CruxCard: &cards[0]}
	} else if u.IsDui(cards) {
		cardType = CARDTYPE_DUI
		cardTypeStruct = &CardTypeStruct{CruxCard: &cards[0]}
	} else if ok, cruxCard := u.IsSanDaiYi(cards); ok {
		cardType = CARDTYPE_SANDAIYI
		cardTypeStruct = &CardTypeStruct{CruxCard: cruxCard}
	} else if ok, cruxCard := u.IsSanDaiEr(cards); ok {
		cardType = CARDTYPE_SANDAIER
		cardTypeStruct = &CardTypeStruct{CruxCard: cruxCard}
	} else if u.IsShunZi(cards) {
		cardType = CARDTYPE_SHUNZI
		cardTypeStruct = &CardTypeStruct{CruxCard: &cards[0]}
	} else if u.IsLianDui(cards) {
		cardType = CARDTYPE_LIANDUI
		cardTypeStruct = &CardTypeStruct{CruxCard: &cards[0]}
	} else if ok, cruxCard := u.IsFeiJi(cards); ok {
		cardType = CARDTYPE_FEIJI
		cardTypeStruct = &CardTypeStruct{CruxCard: cruxCard}
	} else if u.IsSan(cards) {
		cardType = CARDTYPE_SAN
		cardTypeStruct = &CardTypeStruct{CruxCard: &cards[0]}
	} else if u.IsZhaDan(cards) {
		cardType = CARDTYPE_ZHADAN
		cardTypeStruct = &CardTypeStruct{CruxCard: &cards[0]}
	} else if u.IsHuoJian(cards) {
		cardType = CARDTYPE_HUOJIAN
	} else if ok, cruxCard := u.IsSiDaiEr(cards); ok {
		cardType = CARDTYPE_SIDAIER
		cardTypeStruct = &CardTypeStruct{CruxCard: cruxCard}
	}
	return cardType, cardTypeStruct
}

//火箭
func (u *User) IsHuoJian(cards []Card) bool {
	if len(cards) == 2 {
		if cards[0].Priority == Priority_BKing && cards[1].Priority == Priority_SKing {
			return true
		}
	}
	return false
}

//炸弹
func (u *User) IsZhaDan(cards []Card) bool {
	if len(cards) == 4 {
		if cards[0].Priority == cards[1].Priority && cards[0].Priority == cards[2].Priority && cards[0].Priority == cards[3].Priority {
			return true
		}
	}
	return false
}

//单牌
func (u *User) IsDan(cards []Card) bool {
	if len(cards) == 1 {
		return true
	}
	return false
}

//对牌
func (u *User) IsDui(cards []Card) bool {
	if len(cards) == 2 {
		if cards[0].Priority == cards[1].Priority {
			return true
		}
	}
	return false
}

//三牌
func (u *User) IsSan(cards []Card) bool {
	if len(cards) == 3 {
		if cards[0].Priority == cards[1].Priority && cards[0].Priority == cards[2].Priority {
			return true
		}
	}
	return false
}

//三带一
func (u *User) IsSanDaiYi(cards []Card) (bool, *Card) {
	var cruxCard *Card //关键牌
	if len(cards) == 4 {
		cardMap := make(map[int]int)
		for _, card := range cards {
			cardMap[card.Priority] += 1
		}
		if len(cardMap) == 2 {
			oneOfCardPriority := cards[0].Priority
			oneOfCardCount := cardMap[oneOfCardPriority]
			if oneOfCardCount == 1 || oneOfCardCount == 3 {
				if oneOfCardCount == 1 {
					cruxCard = &cards[1]
				} else {
					cruxCard = &cards[0]
				}
				return true, cruxCard
			}
		}
	}
	return false, cruxCard
}

//三带二
func (u *User) IsSanDaiEr(cards []Card) (bool, *Card) {
	var cruxCard *Card //关键牌
	if len(cards) == 5 {
		cardMap := make(map[int]int)
		for _, card := range cards {
			cardMap[card.Priority] += 1
		}
		if len(cardMap) == 2 {
			oneOfCardPriority := cards[0].Priority
			oneOfCardCount := cardMap[oneOfCardPriority]
			if oneOfCardCount == 2 || oneOfCardCount == 3 {
				if oneOfCardCount == 2 {
					cruxCard = &cards[2]
				} else {
					cruxCard = &cards[0]
				}
				return true, cruxCard
			}
		}
	}
	return false, cruxCard
}

//顺子
func (u *User) IsShunZi(cards []Card) bool {
	if len(cards) >= 5 {
		priority := 0
		for _, card := range cards {
			if card.Priority == Priority_BKing || card.Priority == Priority_SKing || card.Priority == Priority_Two {
				return false
			}
			if priority == 0 {
				priority = card.Priority
			} else {
				if priority == card.Priority+1 {
					priority = card.Priority
				} else {
					return false
				}
			}
		}
		return true
	}
	return false
}

//连对
func (u *User) IsLianDui(cards []Card) bool {
	if len(cards) >= 6 && len(cards)%2 == 0 {
		cardMap := make(map[int]int)
		for _, card := range cards {
			if card.Priority == Priority_BKing || card.Priority == Priority_SKing || card.Priority == Priority_Two {
				return false
			}
			cardMap[card.Priority] += 1
		}
		cardPrioritys := []int{}
		for cardPriority, cardCount := range cardMap {
			if cardCount != 2 {
				return false
			}
			cardPrioritys = append(cardPrioritys, cardPriority)
		}
		sort.Ints(cardPrioritys)
		priority := 0
		for _, cardPriority := range cardPrioritys {
			if priority == 0 {
				priority = cardPriority
			} else {
				if priority+1 == cardPriority {
					priority = cardPriority
				} else {
					return false
				}
			}
		}
		return true
	}
	return false
}

//飞机(带翅膀和不带翅膀)
func (u *User) IsFeiJi(cards []Card) (bool, *Card) {
	var cruxCard *Card //关键牌
	if len(cards) >= 6 {
		isHuojian := -1
		cardMap := make(map[int]int)
		for _, card := range cards {
			cardMap[card.Priority] += 1
			if card.Priority > Priority_Two {
				isHuojian += 1
			}
			if cardMap[card.Priority] == 3 && cruxCard == nil {
				cruxCard = card.clone()
			}
		}
		//不能带火箭
		if isHuojian == 1 {
			return false, nil
		}
		oneTwoCardPrioritys := []int{}
		threeCardPrioritys := []int{}
		for cardPriority, cardCount := range cardMap {
			if cardCount == 1 || cardCount == 2 {
				oneTwoCardPrioritys = append(oneTwoCardPrioritys, cardPriority)
			} else if cardCount == 3 {
				threeCardPrioritys = append(threeCardPrioritys, cardPriority)
			}
		}
		sort.Ints(threeCardPrioritys)
		feijiCount := 0
		priority := 0
		for _, cardPriority := range threeCardPrioritys {
			if priority == 0 {
				priority = cardPriority
				feijiCount += 1
			} else {
				if priority+1 == cardPriority {
					priority = cardPriority
					feijiCount += 1
				} else {
					if feijiCount <= 1 {
						priority = cardPriority
						feijiCount = 1
					} else {
						break
					}
				}
			}
		}
		chibangCardCount := 0
		for _, cardCount := range cardMap {
			if cardCount < 3 {
				chibangCardCount += cardCount
			}
		}
		//除去飞机剩余牌中三张一样牌型的数量
		theCount := len(threeCardPrioritys) - feijiCount
		if theCount > 0 {
			chibangCardCount += theCount * 3
		}
		if feijiCount >= 2 {
			//飞机数量等于翅膀的包含的牌数
			if chibangCardCount == 0 {
				return true, cruxCard
			} else if feijiCount == chibangCardCount {
				return true, cruxCard
			} else if feijiCount == chibangCardCount/2 {
				return true, cruxCard
			}
		}
	}
	return false, nil
}

//四带二
func (u *User) IsSiDaiEr(cards []Card) (bool, *Card) {
	var cruxCard *Card //关键牌
	if len(cards) == 6 || len(cards) == 8 {
		cardMap := make(map[int]int)
		for _, card := range cards {
			cardMap[card.Priority] += 1
			if cardMap[card.Priority] == 4 {
				cruxCard = card.clone()
			}
		}
		if len(cardMap) != 2 && len(cardMap) != 3 {
			return false, cruxCard
		}
		counts := []int{}
		for _, cardCount := range cardMap {
			counts = append(counts, cardCount)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(counts)))
		if counts[0] != 4 {
			return false, cruxCard
		}
		if counts[1] != 1 && counts[1] != 2 {
			return false, cruxCard
		}
		return true, cruxCard
	}
	return false, cruxCard
}
