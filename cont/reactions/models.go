package reactions

import (
	"encoding/json"
	"time"
)

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

type UserCachedContents struct {
	UserId                uint32    `json:"user_id" gorm:"primary_key"`
	RecommendedArticles   string    `json:"recommended_articles"`
	RecommendedArticlesAt time.Time `json:"recommended_articles_at"`
}

type RecommendedArticle struct {
	WithUUiDRow
	Score int `json:"s"`
}

type RecommendedArticles map[string]*RecommendedArticle

type WithUUiDRow struct {
	ID string `json:"id" sql:"type:uuid;primary_key;default:uuid_generate_v4()"`
}

func (ucc *UserCachedContents) SetRecommendedArticles(in RecommendedArticles, syncDate bool) error {
	txt, err := json.Marshal(in)
	if err != nil {
		return err
	}

	ucc.RecommendedArticles = string(txt)
	if syncDate {
		ucc.RecommendedArticlesAt = time.Now()
	}
	return nil
}

func (ucc *UserCachedContents) ParseRecommendedArticles() (RecommendedArticles, error) {
	res := make(RecommendedArticles)
	if ucc.RecommendedArticles == "" {
		//empty
		return res, nil
	}

	err := json.Unmarshal([]byte(ucc.RecommendedArticles), &res)
	if err != nil {
		return res, err
	}
	return res, nil
}
