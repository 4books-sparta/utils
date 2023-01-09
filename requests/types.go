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
