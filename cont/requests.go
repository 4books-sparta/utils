package cont

type SkillBooksMatchRequest struct {
	Lang  string   `json:"-" validate:"required"`
	Books []string `json:"books" validate:"required"`
}

type SkillsBooksMatchRequest struct {
	Lang   string   `json:"-" validate:"required"`
	Skills []uint32 `json:"skills" validate:"required"`
}

const (
	TypeBook      = "book"
	TypeArticle   = "article"
	TypePodcast   = "podcast"
	TypeTheUpdate = "theUpdate"
	TypeSkill     = "skill"
	TypeCategory  = "category"
)
