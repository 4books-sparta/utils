package intercom

import (
	"time"

	"gopkg.in/intercom/intercom-go.v2"
)

type ABVariant struct {
	Text      string
	CreatedAt time.Time
}

type User struct {
	Id             string
	Email          string
	Verified       bool              `json:"verified,omitempty"`
	UserVerified   bool              `json:"user_verified,omitempty"`
	FullName       string            `json:"full_name,omitempty"`
	CustomFields   map[string]string `json:"custom_fields,omitempty"`
	CreatedAt      *time.Time        `json:"created_at,omitempty"`
	Subscription   *Subscription     `json:"subscription,omitempty"`
	BooksCompleted uint32            `json:"books_completed,omitempty"`
	BooksStarted   uint32            `json:"books_started,omitempty"`
	LastStarted    *BookActivity     `json:"last_started,omitempty"`
	LastCompleted  *BookActivity     `json:"last_completed,omitempty"`
	Score          *int              `json:"score,omitempty"`
	ABTestVariant  *ABVariant        `json:"ab_test_variant,omitempty"`

	RawUser *intercom.User `json:"raw,omitempty"`
}

type BookActivity struct {
	Slug  string
	Title string
	At    time.Time
}

type Subscription struct {
	Status         string
	Company        string
	Provider       string
	Plan           string
	Products       string
	TrialStart     time.Time
	TrialEnd       time.Time
	Expiry         time.Time
	CreatedAt      time.Time
	CancelledAt    time.Time
	LastDisabledAt time.Time
	LastEnabledAt  time.Time
}

type Response struct {
	User         *User                  `json:"user"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

type SimpleEvent struct {
	EventName string                 `json:"event_name"`
	CreatedAt int64                  `json:"created_at"`
	UserId    string                 `json:"user_id"`
	Email     string                 `json:"email"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
