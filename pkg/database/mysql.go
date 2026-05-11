package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// NewMySQL opens a MySQL connection using the given DSN, configures the
// connection pool, and verifies the connection with a Ping.
func NewMySQL(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("database: failed to open MySQL connection: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(3 * time.Minute)

	// Verify the connection is actually reachable.
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database: gagal ping MySQL: %w", err)
	}

	return db, nil
}
