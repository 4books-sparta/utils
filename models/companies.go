package models

import (
	"time"
)

type Company struct {
	Id                  uint32                         `json:"id"`
	Name                string                         `json:"name" validate:"required"`
	Logo                string                         `json:"logo,omitempty"`
	Slug                string                         `json:"slug" validate:"required,slug"`
	CompanyTranslations map[string]*CompanyTranslation `json:"translations" gorm:"-"`
}

type CompanyTranslation struct {
	Logo      string `json:"logo,omitempty"`
	Claim     string `json:"claim"`
	CompanyId uint32 `json:"company_id"`
}

type CompanyTeam struct {
	Id        uint32 `json:"id"`
	Name      string `json:"name" validate:"required"`
	CompanyId uint32 `json:"company_id" validate:"required"`
}

type Subscription struct {
	Id     uint32
	Expiry time.Time
}

type CompanyUser struct {
	CompanyId      uint32        `json:"company_id"`
	TeamId         uint32        `json:"team_id"`
	UserId         uint32        `json:"user_id"`
	DateIn         *time.Time    `json:"date_in"`
	DateOut        *time.Time    `json:"date_out"`
	SubscriptionId *uint32       `json:"subscription_id"`
	Subscription   *Subscription `json:"subscription,omitempty"`
	User           *User         `json:"user,omitempty"`
	TaskId         *uint32       `json:"task_id"`
}

type CompanyUserFilter struct {
	Limit     uint32 `json:"limit"`
	Offset    uint32 `json:"offset"`
	Email     string `json:"email"`
	CompanyId uint32 `json:"company_id"`
}

type WebRedirect struct {
	Id      uint32 `json:"id"`
	FromUrl string `json:"from_url" validate:"required"`
	ToUrl   string `json:"to_url" validate:"required"`
	Active  bool   `json:"active"`
	Code    int    `json:"code" validate:"oneof=301 302"`
}
