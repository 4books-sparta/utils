package kafka_messages

import (
	"time"

	"github.com/4books-sparta/utils/intercom"
)

const (
	EventBookStarted             = "bs"
	EventBookCompleted           = "bc"
	EventUserUpdated             = "uu"
	EventCategoryFeedbackChanged = "cfc"
	EventSkillFollowChanged      = "sfc"
	EventSubscriptionUpdated     = "su"
	EventFreemiumCreated         = "fs"
	EventSimple                  = "se"
	EventEmailChanged            = "ec"
	EventCompanyUpdated          = "cu"
	EventUser2Company            = "u2c"
	ObjectTypeCompany            = "cmp"
	EventCouponAssigned          = "uca"
	EventCouponConsumed          = "ucc"
	EventLeadTagged              = "lt"
	EventTheUpdateEmailOpened    = "teo"
	EventSameUser                = "sus"
	EventPsychoWebhook           = "psy"
	KeyNum                       = "num"
	MetaKeyEvent                 = "event"
)

func NewIntercomSimpleEvent(userId uint32, ev intercom.SimpleEvent) *IntercomEvent {
	data := make(map[string]interface{})
	data[MetaKeyEvent] = ev
	return &IntercomEvent{
		UserId:    userId,
		EventName: EventSimple,
		Data:      data,
	}
}

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

type PsychoStats struct {
	FreeAppData   *time.Time `json:"free_app_data,omitempty"`
	NextPaidApp   *time.Time `json:"next_paid_app,omitempty"`
	NextAppStatus string     `json:"next_app_status,omitempty"`
	PaidAppNum    uint32     `json:"paid_app_num,omitempty"`
}

func (ev *IntercomEvent) FillPsychoData(stats PsychoStats) {
	data := make(map[string]interface{})
	if stats.PaidAppNum > 0 {
		data["psycho_paid_app"] = stats.PaidAppNum
	}

	if stats.FreeAppData != nil {
		data["psycho_free_app"] = *stats.FreeAppData
	}

	if stats.NextPaidApp != nil {
		data["psycho_next_app"] = *stats.NextPaidApp
	}

	if stats.NextAppStatus != "" {
		data["psycho_payment_status"] = stats.NextAppStatus
	}
	ev.Data = data
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
		//JSON decode treats numbers as float64
		if intV, ok2 := num.(float64); ok2 {
			return uint(intV)
		}
	}
	return 0
}
