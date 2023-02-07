package cont

type SkillBooksMatchRequest struct {
	Lang  string   `json:"-" validate:"required"`
	Books []string `json:"books" validate:"required"`
}
