package mysql

import (
	"database/sql"
	"errors"

	"github.com/alekslesik/file-cloud/pkg/models"
)

type FileModel struct {
	DB *sql.DB
}

// Add a new record to the users table.
func (m *FileModel) Insert(name, fileType string, size int64) (int, error) {
	// SQL request we wanted to execute
	stmt := `INSERT INTO files (name, type, size, created) VALUES(?, ?, ?, UTC_TIMESTAMP())`

	// Use Exec() for execute SQL request
	result, err := m.DB.Exec(stmt, name, fileType, size)
	if err != nil {
		return 0, err
	}

	// Get the last created snippet ID from snippets table
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Return snippet data by ID
func (m *FileModel) Get(id int) (*models.File, error) {
	// SQL request for getting data of one record
	stmt := `SELECT id, name, type, size, created FROM files WHERE id = ?`

	// Use QueryRow() for executing SQL request passing unreliable variable ID like a placeholder
	row := m.DB.QueryRow(stmt, id)

	// Initialise the pointer to new struct Snippet
	s := &models.File{}

	// Use row.Scan() to copy the value from every sql.Row field to Snippet Struct
	err := row.Scan(&s.ID, &s.Name, &s.Type, &s.Size, &s.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}

	// If all ok return Snippet object
	return s, nil
}

// Return all files
func (m *FileModel) All() ([]*models.File, error) {
	// SQL request we wanted to execute
	stmt := `SELECT id, name, type, size, created FROM files ORDER BY created`

	// Use Query() for execute SQL request
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var files []*models.File

	// Use rows.Next() to run over the result
	for rows.Next() {
		s := &models.File{}
		// Use row.Scan() to copy the value from every sql.Row field to File Struct
		err = rows.Scan(&s.ID, &s.Name, &s.Type, &s.Size, &s.Created)
		if err != nil {
			return nil, err
		}
		// Add the struct to slice
		files = append(files, s)
	}

	// Call rows.Err() after rows.Next() to ensure we haven't any errors
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// If all ok return slice
	return files, nil
}
