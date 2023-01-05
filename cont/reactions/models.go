package reactions

import "time"

type Feedback struct {
	UserId    uint32     `json:"user_id,omitempty"  validate:"required" gorm:"primary_key"`
	Locale    string     `json:"-" gorm:"primary_key"`
	Liked     int8       `json:"liked"`
	CreatedAt *time.Time `json:"-"`
	UpdatedAt time.Time  `json:"-"`
}

type SkillFeedback struct {
	SkillId uint32 `json:"skill_id"  validate:"required" gorm:"primary_key"`
	Feedback
}

type ContFavourite struct {
	UserId    uint32     `json:"user_id,omitempty"  validate:"required" gorm:"primary_key"`
	Locale    string     `json:"-" gorm:"primary_key"`
	CreatedAt *time.Time `json:"-"`
}

type SkillFavourite struct {
	SkillId uint32 `json:"skill_id"  validate:"required" gorm:"primary_key"`
	ContFavourite
}

func (f *SkillFeedback) Essentials() *SkillFeedback {
	ret := SkillFeedback{
		SkillId: f.SkillId,
		Feedback: Feedback{
			Liked: f.Liked,
		},
	}
	return &ret
}
