package kafka_messages

import "time"

const (
	OnUserActionSyncKPI = "sync_kpis"
)

type ProgressEvent struct {
	Locale         string    `json:"l"`
	Ts             time.Time `json:"t"`
	UserId         uint32    `json:"u"`
	Seconds        uint32    `json:"s"`
	AudioTot       uint32    `json:"a"`
	Chapter        uint32    `json:"c"`
	MaxChapters    uint32    `json:"m"`
	Completed      bool      `json:"cd"`
	EverCompleted  bool      `json:"e"`
	FirstCompleted bool      `json:"f"`
}

type BookProgressEvent struct {
	ProgressEvent
	BookId string `json:"b"`
}

type OnUserActionEvent struct {
	UserId uint32            `json:"u"`
	Action string            `json:"a"`
	Params map[string]string `json:"p"`
	Ts     time.Time         `json:"t"`
}

type SkillFeedbackEvent struct {
	UserId  uint32    `json:"u"`
	SkillId uint32    `json:"s"`
	Locale  string    `json:"l"`
	Action  int8      `json:"a"`
	Ts      time.Time `json:"t"`
}

type UserFunnelEvent struct {
	UserId      uint32            `json:"u" validate:"required"`
	EventType   string            `json:"e" validate:"required"`
	Platform    string            `json:"p" validate:"required"`
	Ts          time.Time         `json:"ts" validate:"required"`
	UtmSource   string            `json:"utms,omitempty"`
	UtmMedium   string            `json:"utmm,omitempty"`
	UtmCampaign string            `json:"utmc,omitempty"`
	UtmContent  string            `json:"utmct,omitempty"`
	UtmTerm     string            `json:"utmt,omitempty"`
	Extras      map[string]string `json:"ext,omitempty"`
}
