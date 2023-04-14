package models

type User struct {
	Id        uint32 `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type UsersFilter struct {
	User             *User
	StrictUserFilter bool
	Limit            uint32
	Offset           uint32
	OrderBy          UsersOrderByFilter
	Order            int8
	ExtSubId         string
}

type UsersOrderByFilter string
