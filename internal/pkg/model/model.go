package model

import (
	"database/sql"

	"github.com/alekslesik/file-cloud/pkg/models"
	"github.com/alekslesik/file-cloud/pkg/models/mysql"
)

func New(db *sql.DB) *Model {
	return &Model{
		Files: &mysql.FileModel{DB: db},
		Users: &mysql.UserModel{DB: db},
	}
}

type Model struct {
	Files interface {
		Insert(name, fileType string, size int64) (int, error)
		Get(id int) (*models.File, error)
		All() ([]*models.File, error)
	}
	Users interface {
		Insert(name, email, password string) error
		Authenticate(email, password string) (int, string, error)
		Get(id int) (*models.User, error)
	}
}
