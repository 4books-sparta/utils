package --service-name--

import (

	"github.com/4books-sparta/utils"
	"gorm.io/gorm"

)

type Repo interface {
}

type SqlRepo struct {
	db      *utils.SqlDatabase
	Testing bool
}

func NewSqlRepo(db *utils.SqlDatabase) SqlRepo {
	return SqlRepo{db: db}
}

func (r SqlRepo) StartTransaction() *gorm.DB {
	if r.Testing {
		return r.db.DB
	}

	return r.db.Begin()
}

func (r SqlRepo) CommitTransaction(tx *gorm.DB) error {
	if r.Testing {
		return nil
	}
	return tx.Commit().Error
}

func (r SqlRepo) RollbackTransaction(tx *gorm.DB) {
	if r.Testing {
		return
	}
	tx.Rollback()
}

