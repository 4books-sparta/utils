package models

import "time"

type Action uint8

const (
	ActionNone = Action(iota)
	ActionCreate
	ActionUpdate
	ActionBuy
	ActionRevoke
)

type Customer struct {
	FirstName string
	LastName  string
	Email     string
	ID        uint32
}

type Effect struct {
	Action                   Action
	Customer                 *Customer
	ExternalId               string
	ExternalPlanId           string
	SubscriptionId           uint32
	Price                    int32
	RequestPlanChange        bool
	Status                   uint8
	Expiry                   *time.Time
	RegisterSubChange        *SubscriptionEventType
	RegisterSubChangeDetails string

	MultiBuy []*EffectDetail
}

type EffectDetail struct {
	ExternalPlanId string
	Price          int32
}
