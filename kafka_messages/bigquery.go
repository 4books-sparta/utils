package kafka_messages

import (
	"errors"
	"time"

	"gopkg.in/go-playground/validator.v9"
)

const (
	ErrorUnknownType            = "unknown-type"
	ErrorWrongDataInterfaceType = "wrong-data-signature"
	BigqueryTypeUserLocation    = "ul"
)

type BigqueryUserLocation struct {
	UserId    uint32     `json:"user_id" validate:"required"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Country   string     `json:"country,omitempty"`
	City      string     `json:"city,omitempty"`
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
	}
	return errors.New(ErrorUnknownType)
}
