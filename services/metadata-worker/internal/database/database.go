package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"metadata-worker/internal/config"
)

const createTable = `CREATE TABLE IF NOT EXISTS metadata_events (
	id BIGSERIAL PRIMARY KEY,
	event_type TEXT NOT NULL,
	project TEXT NOT NULL,
	repository TEXT NOT NULL,
	tag TEXT,
	digest TEXT,
	operator TEXT,
	occurred_at TIMESTAMPTZ NOT NULL,
	received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`

// Open connects to PostgreSQL and initializes the schema.
func Open(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBSSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	if _, err := db.Exec(createTable); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
