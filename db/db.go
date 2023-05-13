package db

import (
	"database/sql"
	"fmt"
	"os"
)

var DB_conn *sql.DB

func Connect() error {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"))

	conn, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	DB_conn = conn
	return nil
}
