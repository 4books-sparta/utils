package models

import "time"

type Plan struct {
	Id              uint32           `json:"id" gorm:"primary_key"`
	Slug            string           `json:"name"`
	Price           uint32           `json:"price"`
	Period          uint8            `json:"period"`
	PeriodUnit      string           `json:"period_unit"`
	Products        []*Product       `json:"products" gorm:"many2many:plan_products;"`
	TrialPeriod     uint8            `json:"trial_period"`
	TrialPeriodUnit string           `json:"trial_period_unit"`
	Status          uint8            `json:"status"`
	Current         *PlanTranslation `json:"translation" gorm:"-"`
}

type PlanUpgrade struct {
	Id                      uint32                  `json:"id" gorm:"primary_key"`
	PlanFromId              uint32                  `json:"-"`
	PlanToId                uint32                  `json:"-"`
	PlanFrom                *Plan                   `json:"plan_from"`
	PlanTo                  *Plan                   `json:"plan_to"`
	Ord                     uint8                   `json:"ord"`
	PlanUpgradesTranslation *PlanUpgradeTranslation `json:"translation"`
}

type PlanUpgradeTranslation struct {
	PlanUpgradeId uint32 `json:"-" gorm:"primary_key"`
	Locale        string `json:"-" gorm:"primary_key"`
	Title         string `json:"title"`
	Subtitle      string `json:"subtitle"`
	ButtonText    string `json:"button_text"`
	ButtonClaim   string `json:"button_claim"`
}

type PlanTranslation struct {
	PlanId          uint32 `json:"-" gorm:"primary_key"`
	Locale          string `json:"-" gorm:"primary_key"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Label           string `json:"label"`
	Promo           string `json:"promo"`
	FullDescription string `json:"full_description"`
	Price           string `json:"price"`
	Duration        string `json:"duration"`
	PromoExtended   string `json:"promo_extended"`
}

type PlanProvider struct {
	Id   uint32 `json:"id" gorm:"primary_key"`
	Name string `json:"name" validate:"required"`
}

type PlanProvision struct {
	PlanId         uint32       `json:"-" gorm:"primary_key"`
	Plan           Plan         `json:"plan"`
	PlanProviderId uint32       `json:"-" gorm:"primary_key"`
	PlanProvider   PlanProvider `json:"plan_provider"`
	ExternalId     string       `json:"external_id"`
}

type Subscription struct {
	Id             uint32           `json:"id" gorm:"primary_key"`
	UserId         uint32           `json:"user_id"`
	PlanId         uint32           `json:"plan_id"`
	Plan           Plan             `json:"plan"`
	PlanProviderId uint32           `json:"-"`
	PlanProvider   PlanProvider     `json:"plan_provider,omitempty"`
	ExternalId     string           `json:"external_id"`
	Expiry         time.Time        `json:"expiry"`
	ExpiryAt       CompleteDatetime `json:"expiry_at,omitempty" gorm:"-"`
	Status         uint8            `json:"status"`
	CreatedAt      time.Time        `json:"-"`
	SubStart       time.Time        `json:"start,omitempty" gorm:"-"`
	InTrial        bool             `json:"in_trial,omitempty" gorm:"-"`
	Trial          *Trial           `json:"trial,omitempty" gorm:"-"`
}

type SubscriptionEvent struct {
	Id             uint32                `json:"id" gorm:"primary_key"`
	SubscriptionId uint32                `json:"subscription_id" gorm:"primary_key"`
	EventType      SubscriptionEventType `json:"event_type"`
	Details        string                `json:"details"`
	CreatedAt      time.Time             `json:"created_at"`
}

type Plangroup struct {
	Id        uint32 `json:"id" gorm:"primary_key"`
	Slug      string `json:"slug"`
	IsDefault uint8  `json:"is_default"`
}

type Referral struct {
	Id         uint32    `json:"id" gorm:"primary_key"`
	ReferredId uint32    `json:"id_referred"`
	ReferrerId uint32    `json:"id_referrer"`
	CreatedAt  time.Time `json:"-"`
}

func (r *Referral) GetUserByRecipient(recipient string) uint32 {
	if recipient == RecipientReferrer {
		return r.ReferrerId
	}

	return r.ReferredId
}

type ReferralBonus struct {
	DateFrom       time.Time `json:"date_from"`
	DateTo         time.Time `json:"date_to"`
	Recipient      string    `json:"payee"`
	TriggerEvent   string    `json:"event"`
	BonusType      string    `json:"bonus"`
	BonusQuantity  int       `json:"value"`
	BonusUnit      string    `json:"unit"`
	PlanProviderId *uint32   `json:"provider_id"`
	PlanSlug       string    `json:"plan"`
}

type ReferralBonusCredit struct {
	Id             uint32        `json:"id" gorm:"primary_key"`
	CreatedAt      time.Time     `json:"-"`
	ReferralId     uint32        `json:"referral_id"`
	Referral       *Referral     `json:"-" gorm:"-"`
	UserId         uint32        `json:"user_id"`
	Recipient      string        `json:"payee"`
	TriggerEvent   string        `json:"event"`
	BonusType      string        `json:"bonus"`
	BonusQuantity  int           `json:"value"`
	BonusUnit      string        `json:"unit"`
	Redeemed       bool          `json:"redeemed"`
	RedeemedAt     *time.Time    `json:"-"`
	SubscriptionId *uint32       `json:"-"`
	Subscription   *Subscription `json:"subscription" gorm:"-"`
	PlanSlug       string        `json:"plan"`
}

type SubscriptionAddon struct {
	SubscriptionId uint32 `json:"-" gorm:"primary_key"`
	ProductId      uint32 `json:"-"`
}

type PlanFilter struct {
	GroupName    string
	ProviderName string
}

type Trial struct {
	SubscriptionId uint32    `gorm:"primary_key"`
	Start          time.Time `gorm:"column:start_time"`
	End            time.Time `gorm:"column:end_time"`
}

func (t *Trial) Status() TrialStatus {
	now := time.Now()

	if !t.Start.After(now) && t.End.After(now) {
		return TrialActive
	}

	if t.End.Before(now) {
		return TrialCompleted
	}

	return TrialVoid
}

func (t *Trial) Active() bool {
	return t.Status() == TrialActive
}

type FreemiumSubscription struct {
	UserId uint32 `gorm:"primary_key"`
	From   time.Time
	To     time.Time
}

func (fs FreemiumSubscription) ToFreemiumResponse() FreemiumResponse {
	return FreemiumResponse{
		From: fs.From.Unix(),
		To:   fs.To.Unix(),
	}
}

type FreemiumResponse struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

type TrialStatus uint8

const (
	Created uint8 = iota
	Active
	Inactive
)

const (
	TrialVoid = TrialStatus(iota)
	TrialActive
	TrialCompleted
)

type SubscriptionEventType uint8

const (
	EventTypeSubscriptionRemoveCancel    = SubscriptionEventType(1)
	EventTypeSubscriptionDeactivation    = SubscriptionEventType(2)
	EventTypeSubscriptionCancel          = SubscriptionEventType(3)
	EventTypeSubscriptionExtend          = SubscriptionEventType(4)
	EventTypeSubscriptionPlanChange      = SubscriptionEventType(5)
	EventTypeSubscriptionBillingIssue    = SubscriptionEventType(6)
	EventTypeSubscriptionBillIssueAutoOk = SubscriptionEventType(7)
)
