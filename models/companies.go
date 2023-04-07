package models

import "time"

type Company struct {
	Id   uint32 `json:"id"`
	Name string `json:"name" validate:"required"`
	Slug string `json:"slug" validate:"required,slug"`
}

type CompanyUser struct {
	CompanyId      uint32            `json:"company_id"`
	UserId         uint32            `json:"user_id"`
	DateIn         *time.Time        `json:"date_in"`
	DateOut        *time.Time        `json:"date_out"`
	SubscriptionId *uint32           `json:"subscription_id"`
	Subscription   *sub.Subscription `json:"subscription,omitempty"`
	User           *User             `json:"user,omitempty"`
	TaskId         *uint32           `json:"task_id"`
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
