package stats

import (
	"fmt"
	"strconv"
	"time"
)

type TimeSerie struct {
	Time  time.Time `gorm:"-"`
	Day   uint32    `json:"day"`   //20231231
	Year  uint32    `json:"year"`  //2023
	Month uint32    `json:"month"` //202301
	Num   uint32    `json:"num"`
}

func (ts *TimeSerie) FillDay() uint32 {
	mon := ts.FillMonth()
	t := fmt.Sprintf("%d%s", mon, ts.Time.Format("02"))
	toInt, _ := strconv.Atoi(t)
	ts.Day = uint32(toInt)
	return ts.Day
}

func (ts *TimeSerie) FillYear() uint32 {
	ts.Year = uint32(ts.Time.Year())
	return ts.Year
}

func (ts *TimeSerie) FillMonth() uint32 {
	y := ts.FillYear()
	t := fmt.Sprintf("%d%s", y, ts.Time.Format("01"))
	toInt, _ := strconv.Atoi(t)
	ts.Month = uint32(toInt)
	return ts.Month
}

type LocalizedTimeSerie struct {
	TimeSerie
	Locale string `json:"lang"`
}
