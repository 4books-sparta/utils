package requests

type StringIdRequest struct {
	Id string `json:"id" validate:"required"`
}

type IdRequest struct {
	Id int `json:"id" validate:"required"`
}

type PagedRequest struct {
	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}

type UserSearchRequest struct {
	UserId uint32 `json:"user_id"`
}

type LocalizedRequest struct {
	Locale string `json:"locale" validate:"required"`
}

type LocalizedPagedUserSearchRequest struct {
	PagedRequest
	UserSearchRequest
	LocalizedRequest
}

type ContRequest struct {
	ContType      string   `json:"cont_type"`
	PublishStates []string `json:"publish_states"`
	LocalizedPagedUserSearchRequest
	IntId    int    `json:"id"`
	StringId string `json:"s_id"`
	Slug     string `json:"slug" validate:"slug"`
}

type SlugRequest struct {
	Slug string `json:"slug"`
}
