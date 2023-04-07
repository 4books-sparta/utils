package models

type ProductType string

const (
	TypeBook       = ProductType("book")
	Type4Books     = ProductType("4books")
	TypeTheUpdate  = ProductType("theupdate")
	TypeArticle    = ProductType("article")
	ListenComplete = 0.90
)

type Product struct {
	Id   uint32      `json:"id" gorm:"primary_key"`
	Slug ProductType `json:"slug"`
}
