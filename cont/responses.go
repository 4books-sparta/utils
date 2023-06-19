package cont

type SkillBooksMatchResponse struct {
	Matches map[uint32]int `json:"matches"`
}

type BooksMatchResponse struct {
	Matches map[int][]string `json:"matches"`
}

type SitemapRow struct {
	ID      string                 `json:"id"`
	Current *SitemapRowTranslation `json:"translation"`
}

type SitemapRows map[string][]*SitemapRow

type SitemapRowTranslation struct {
	Slug string `json:"slug"`
}

type ResourceStats struct {
	Resources int
	Users     int
}
