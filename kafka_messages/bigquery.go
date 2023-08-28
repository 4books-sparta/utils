package kafka_messages

import (
	"errors"
	"time"

	"gopkg.in/go-playground/validator.v9"
)

const (
	ErrorUnknownType                     = "unknown-type"
	ErrorWrongDataInterfaceType          = "wrong-data-signature"
	BigqueryTypeUserLocation             = "ul"
	BigqueryTypeStatPublishedBooks       = "pb"
	BigqueryTypeStatPublishedArticles    = "pa"
	BigqueryTypeStatSignups              = "usu"
	BigqueryTypeSearchesLog              = "slo"
	BigqueryTypeStatSubscriptions        = "u_subs"
	BigqueryTypeStatCreatedSubscriptions = "t_sub"
	BigqueryTypeStatActiveSubscriptions  = "a_sub"
	BigqueryTypeStatActiveTrials         = "a_try"
	BigqueryTypeStatExpiresSub           = "e_sub"
	BigqueryTypePublishedSkills          = "psk"
	BigqueryTypePublishedMagazinePosts   = "pmp"
	BigquerySubscriptionType             = "sub"
	BigqueryFreeTrialType                = "free_trial"
	BigqueryNoTrialType                  = "no_trial"
)

type BigqueryDailySignups struct {
	Date     string `json:"date,omitempty"`
	Year     uint16 `json:"year,omitempty"`
	Month    uint8  `json:"month,omitempty"`
	Day      uint8  `json:"day,omitempty"`
	Lang     string `json:"lang"`
	Num      uint   `json:"num"`
	Platform string `json:"platform"`
}

type BigqueryDailySubs struct {
	Date  string `json:"date,omitempty"`
	Year  uint16 `json:"year,omitempty"`
	Month uint8  `json:"month,omitempty"`
	Day   uint8  `json:"day,omitempty"`
	// sub | free_trial | no_trial
	Type      string `json:"type,omitempty"`
	Num       uint   `json:"num"`
	Platform  string `json:"platform,omitempty"`
	Plan      string `json:"plan"`
	Provision string `json:"provision"`
}

type BigqueryUserLocation struct {
	UserId    uint32     `json:"user_id" validate:"required"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Country   string     `json:"country,omitempty"`
	City      string     `json:"city,omitempty"`
}

type BigquerySearchesLog struct {
	Ts      int    `json:"ts"`
	Type    string `json:"type"`
	Qs      string `json:"qs"`
	Num     int    `json:"num"`
	Results string `json:"results"`
}

type BigqueryMsg struct {
	Type string
	Data interface{}
	Log  func(string) `json:"-"`
}

func (msg *BigqueryMsg) Validate() error {
	validate := validator.New()

	switch msg.Type {
	case BigqueryTypeUserLocation:
		if _, ok := msg.Data.(BigqueryUserLocation); !ok {
			if msg.Log != nil {
				msg.Log(ErrorWrongDataInterfaceType)
			}
			return errors.New(ErrorWrongDataInterfaceType)
		}
		//Validate struct
		return validate.Struct(msg.Data)
	case BigqueryTypeStatSignups:
		if _, ok := msg.Data.(BigqueryDailySignups); !ok {
			if msg.Log != nil {
				msg.Log(ErrorWrongDataInterfaceType)
			}
			return errors.New(ErrorWrongDataInterfaceType)
		}
		//Validate struct
		return validate.Struct(msg.Data)
	case BigqueryTypeStatSubscriptions:
		if _, ok := msg.Data.(BigqueryDailySubs); !ok {
			if msg.Log != nil {
				msg.Log(ErrorWrongDataInterfaceType)
			}
			return errors.New(ErrorWrongDataInterfaceType)
		}
		//Validate struct
		return validate.Struct(msg.Data)
	case BigqueryTypeSearchesLog:
		if _, ok := msg.Data.(BigquerySearchesLog); !ok {
			if msg.Log != nil {
				msg.Log(ErrorWrongDataInterfaceType)
			}
			return errors.New(ErrorWrongDataInterfaceType)
		}
		//Validate struct
		return validate.Struct(msg.Data)
	}
	return errors.New(ErrorUnknownType)
}
