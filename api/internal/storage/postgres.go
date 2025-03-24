package storage

import (
	"fmt"

	"github.com/liriquew/control_system/internal/config"

	"github.com/jmoiron/sqlx"
)

func NewStorage(cfg *config.StorageConfig) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return db, nil
}
