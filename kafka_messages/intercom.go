package kafka_messages

import "time"

const (
	EventBookStarted             = "bs"
	EventBookCompleted           = "bc"
	EventUserUpdated             = "uu"
	EventCategoryFeedbackChanged = "cfc"
	EventSkillFollowChanged      = "sfc"
	EventSubscriptionUpdated     = "su"
	EventFreemiumCreated         = "fs"
	EventEmailChanged            = "ec"
	EventCompanyUpdated          = "cu"
	EventUser2Company            = "u2c"
	ObjectTypeCompany            = "cmp"
	EventCouponAssigned          = "uca"
	EventCouponConsumed          = "ucc"
	EventLeadTagged              = "lt"
	EventTheUpdateEmailOpened    = "teo"
	KeyNum                       = "num"
)

type IntercomEvent struct {
	UserId    uint32                 `json:"u"`
	EventName string                 `json:"e"`
	Locale    string                 `json:"l,omitempty"`
	ItemType  string                 `json:"i,omitempty"`
	ItemId    string                 `json:"id,omitempty"`
	Platform  string                 `json:"p,omitempty"`
	DateStart *time.Time             `json:"ds,omitempty"`
	DateEnd   *time.Time             `json:"de,omitempty"`
	Ts        time.Time              `json:"t"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

func (ev *IntercomEvent) SetTheUpdateEmailOpenedNum(num uint) {
	data := make(map[string]interface{})
	data[KeyNum] = num
	ev.Data = data
}

func (ev *IntercomEvent) GetTheUpdateEmailOpenedNum() uint {
	if ev.Data == nil {
		return 0
	}
	if num, ok := ev.Data[KeyNum]; ok {
		if intV, ok2 := num.(uint); ok2 {
			return intV
		}
	}
	return 0
}
