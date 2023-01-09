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
