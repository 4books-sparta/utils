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
