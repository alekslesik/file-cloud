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
		Insert(fileName string, fileType string, fileSize int64, fileUrl string, userID int) (int, error)
		Get(id int) (*models.File, error)
		All(userId int) ([]*models.File, error)
	}
	Users interface {
		Insert(name, email, password string) error
		Authenticate(email, password string) (int, string, error)
		Get(id int) (*models.User, error)
	}
}
