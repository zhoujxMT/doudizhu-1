/*
牌类
*/

package engine

//牌类型
const (
	CARDTYPE_HUOJIAN = iota
	CARDTYPE_ZHADAN
	CARDTYPE_DAN
	CARDTYPE_DUI
	CARDTYPE_SAN
	CARDTYPE_SANDAIYI
	CARDTYPE_SANDAIER
	CARDTYPE_SHUNZI
	CARDTYPE_LIANDUI
	CARDTYPE_FEIJI
	CARDTYPE_SIDAIER
)

//花色
const (
	SUIT_HEITAO   = 1
	SUIT_HONGTAO  = 2
	SUIT_MEIHUA   = 3
	SUIT_FANGKUAI = 4
)

var (
	CardTypeNames = []string{"火箭", "炸弹", "单", "对", "三", "三带一", "三带二", "顺子", "连对", "飞机", "四带二"}
)

var (
	CardMode = -1
)

func GetCardMode() int {
	return CardMode
}

func SetCardMode(cardMode int) {
	CardMode = cardMode
}

type Card struct {
	//ID(0-53)
	ID int
	//牌值(1-15)
	Value int
	//花色(0:王 1:黑桃 2:红桃 3:梅花 4:方块)
	Suit int
	//优先级（数越大的值约大 1-15）
	Priority int
	//一套牌中的索引
	Index int
}

func (this *Card) clone() *Card {
	return &Card{this.ID, this.Value, this.Suit, this.Priority, this.Index}
}
