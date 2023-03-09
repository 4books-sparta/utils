package reactions

import "time"

const (
	ReactionFollowKey = "follow"
	ReactionLiked     = "liked"
)

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
	Locale    string     `json:"locale,omitempty"  validate:"required" gorm:"primary_key"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type SkillFavouriteSaveRequest struct {
	Rows   []*SkillFavourite `json:"data" validate:"required,dive"`
	Clean  bool              `json:"-"`
	UserId uint32            `json:"-"`
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

type SkillFollowedResponse struct {
	Skills []uint32 `json:"skills"`
}
