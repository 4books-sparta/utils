package intercom

import (
	"time"
)

type ABVariant struct {
	Text      string
	CreatedAt time.Time
}

type User struct {
	Id             string
	Email          string
	Verified       bool
	UserVerified   bool
	FullName       string
	CustomFields   map[string]string
	CreatedAt      time.Time
	Subscription   *Subscription
	BooksCompleted uint32
	BooksStarted   uint32
	LastStarted    BookActivity
	LastCompleted  BookActivity
	Score          *int
	ABTestVariant  *ABVariant
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
