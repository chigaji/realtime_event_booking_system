package database

import (
	"database/sql"
	"time"

	"github.com/chigaji/realtime_event_booking_system/internal/config"
	_ "github.com/lib/pq"
)

func Init(cfg config.DatabaseConfig) (*sql.DB, error) {

	db, err := sql.Open("postgres", cfg.DNS)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(time.Minute * 5)
	return db, nil
}
