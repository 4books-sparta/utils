package cont

type SkillBooksMatchResponse struct {
	Matches map[uint32]int `json:"matches"`
}

type BooksMatchResponse struct {
	Matches map[int][]string `json:"matches"`
}
