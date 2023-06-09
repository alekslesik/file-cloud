package mysql

import (
	"database/sql"
	"strings"

	"github.com/alekslesik/file-cloud/pkg/models"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}

// Add a new record to the users table.
func (m *UserModel) Insert(name, email, password string) error {
	// Create a bcrypt hash of the plain-text password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	// SQL request we wanted to execute.
	stmt := `INSERT INTO users (name, email, hashed_password, created)
    VALUES(?, ?, ?, UTC_TIMESTAMP())`

	// Use the Exec() method to insert the user details and hashed password
	// into the users table. If this returns an error, we try to type assert
	// it to a *mysql.MySQLError object so we can check if the error number is
	// 1062 and, if it is, we also check whether or not the error relates to
	// our users_uc_email key by checking the contents of the message string.
	// If it does, we return an ErrDuplicateEmail error. Otherwise, we just
	// return the original error (or nil if everything worked).
	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			// if mysqlErr.Number == 1062 {
			// 	return models.ErrDuplicateEmail
			// }
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "Duplicate entry") {
				return models.ErrDuplicateEmail
			}
		}
	}

	return err
}

// Verify whether a user exists with the provided email address and password.
// Return the relevant user ID if they do.
func (m *UserModel) Authenticate(email, password string) (int, string, error) {
	var id int
	var name string
	var hashedPassword []byte

	// Retrieve the id and hashed password associated with the given email. If
	// matching email exists, we return the ErrInvalidCredentials error.
	row := m.DB.QueryRow("SELECT id, name, hashed_password FROM users WHERE email = ?", email)
	err := row.Scan(&id, &name, &hashedPassword)
	if err == sql.ErrNoRows {
		return 0, "", models.ErrInvalidCredentials
	} else if err != nil {
		return 0, "", err
	}

	// Check whether the hashed password and plain-text password provided match
	// If they don't, we return the ErrInvalidCredentials error.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", models.ErrInvalidCredentials
	} else if err != nil {
		return 0, "", err
	}

	return id, name, nil
}

// Fetch details for a specific user based on their user ID.
func (m *UserModel) Get(id int) (*models.User, error) {
	s := &models.User{}

	stmt := `SELECT id, name, email, created FROM users WHERE id = ?`
	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Name, &s.Email, &s.Created)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	return s, nil
}
