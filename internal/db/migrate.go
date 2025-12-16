package db

import (
	"database/sql"
	"os"
)

func Migrate(db *sql.DB) error {
	data, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(data))
	return err
}
