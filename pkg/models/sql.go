package models

import (
	"database/sql"
	"errors"
	"log"
)

var ErrNoDriver = errors.New("driver: not supported")

const (
	MYSQL = "mysql"
)

// Open DB connection pool depends on driver
func OpenDB(dsn, driver string) (*sql.DB, error) {
	const op = "helpers.OpenDB()"
	var db *sql.DB
	var err error

	switch driver {
	case MYSQL:
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("%s: open db error: %s", op, err)
			return nil, err
		}
		if err = db.Ping(); err != nil {
			log.Printf("%s: ping db error: %s", op, err)
			return nil, err
		}
	default:
		return nil, ErrNoDriver
	}

	return db, nil
}
