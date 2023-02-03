package helper

import "github.com/4books-sparta/utils"

type Repo interface {
}

type SqlRepo struct {
	db      *utils.SqlDatabase
	Testing bool
}

func NewSqlRepo(db *utils.SqlDatabase) SqlRepo {
	return SqlRepo{db: db}
}
