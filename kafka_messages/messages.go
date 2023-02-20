package kafka_messages

import "time"

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
