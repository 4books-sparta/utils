package game

import "time"

type SkillXP struct {
	SkillId uint32 `json:"skill_id" validate:"required" gorm:"primaryKey"`
	Locale  string `json:"locale,omitempty" validate:"required" gorm:"primaryKey"`
	XP      int    `json:"xp"`
}

type SkillXPRequest struct {
	UserId   uint32     `json:"user_id" validate:"required"`
	SkillXps []SkillXP  `json:"skill_xps" validate:"required,dive"`
	Ts       *time.Time `json:"ts"`
}
