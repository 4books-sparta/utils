package models

import "time"

type User struct {
	Id            uint32    `json:"id" gorm:"primary_key"`
	FirstName     string    `json:"first_name" validate:"required"`
	LastName      string    `json:"last_name" validate:"required"`
	Email         string    `json:"email" validate:"required,email"`
	CreatedAt     time.Time `json:"created_at"`
	EmailVerified bool      `json:"verified"`
	VerifyCode    string    `json:"vc"`
	ReferralCode  string    `json:"referral_code" gorm:"-"`
	UserVerified  bool      `json:"user_verified"`
}

type UsersFilter struct {
	User             *User
	StrictUserFilter bool
	Limit            uint32
	Offset           uint32
	OrderBy          UsersOrderByFilter
	Order            int8
	ExtSubId         string
}

type UsersOrderByFilter string

type Address struct {
	Id       uint32 `json:"id" gorm:"primary_key"`
	UserId   uint32 `json:"user_id" validate:"required"`
	Street   string `json:"street" validate:"required"`
	Postcode string `json:"postcode" validate:"required"`
	City     string `json:"city" validate:"required"`
	State    string `json:"state" validate:"required"`
	Country  string `json:"country" validate:"required"`
}

type UserPref struct {
	UserId uint32 `json:"-" validate:"required" gorm:"primary_key"`
	Key    string `json:"key" validate:"required" gorm:"primary_key"`
	Value  string `json:"value"`
}

type FiscalIdentity struct {
	UserId    uint32 `json:"user_id" gorm:"primary_key"`
	Country   string `json:"country" gorm:"primary_key"`
	Code      string `json:"code"`
	VatNumber string `json:"vat_number"`
}
