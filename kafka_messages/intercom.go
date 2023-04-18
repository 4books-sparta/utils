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
