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

type CompanyFeaturedSkill struct {
	CompanyId uint32 `json:"company_id"`
	TeamId    uint32 `json:"team_id"`
	SkillId   uint32 `json:"skill_id"`
	Ord       int    `json:"ord"`
}

type CompanyFeaturedSkillRequest struct {
	Skill *CompanyFeaturedSkill `json:"skill"`
}

type CompanyFeaturedSkillResponse struct {
	Skill *CompanyFeaturedSkill `json:"skill"`
}

type CompanyFeaturedSkillsResponse struct {
	Skills []*CompanyFeaturedSkill `json:"skills"`
}

func GetSupportedLanguages() []string {
	return []string{"it", "en", "es"}
}

type PerUsersRequests struct {
	Ids []uint32
}
