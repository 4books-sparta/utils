package models

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
