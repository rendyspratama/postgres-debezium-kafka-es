package utils

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	db   *sql.DB
	once sync.Once
)

func GetDB() *sql.DB {
	once.Do(func() {
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://user:password@localhost:5432/digital_discovery?sslmode=disable"
		}

		var err error
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			panic(fmt.Sprintf("Failed to connect to database: %v", err))
		}

		if err := db.Ping(); err != nil {
			panic(fmt.Sprintf("Failed to ping database: %v", err))
		}

		// Set connection pool settings
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
	})

	return db
}
