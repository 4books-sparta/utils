package models

import (
	"fmt"
	"strings"
	"time"
)

const (
	DayLayout     = "2006-01-02"
	DayHourLayout = "2006-01-02T15:04:05Z07:00"
)

type CompleteDatetime struct {
	time.Time
}

func (t *CompleteDatetime) UnmarshalJSON(b []byte) (err error) {
	loaded, err := UnmarshalTime(b, DayHourLayout)
	if err != nil {
		return err
	}

	t.Time = loaded

	return nil
}

func (t *CompleteDatetime) MarshalJSON() ([]byte, error) {
	return MarshalTime(t.Time, DayHourLayout)
}

func (t *CompleteDatetime) MarshalText() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}

	return []byte(t.Format(DayHourLayout)), nil
}

func (t *CompleteDatetime) IsSet() bool {
	return !t.Time.IsZero()
}

func (t *CompleteDatetime) Parse(s string) error {
	raw, err := time.Parse(DayHourLayout, s)
	if err != nil {
		return err
	}

	t.Time = raw
	return nil
}

func UnmarshalTime(b []byte, layout string) (time.Time, error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		return time.Time{}, nil
	}

	return time.Parse(layout, s)
}

func MarshalTime(t time.Time, layout string) ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}

	return []byte(fmt.Sprintf("\"%s\"", t.Format(layout))), nil
}
