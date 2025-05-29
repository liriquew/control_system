package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/liriquew/auth_service/internal/lib/config"
	"github.com/liriquew/auth_service/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrUserExists = errors.New("user already exists")
)

type Repository struct {
	db *sqlx.DB
}

const listTasksBatchSize = 10

func NewAuthRepository(cfg config.StorageConfig) (*Repository, error) {
	const op = "storage.postgres.New"

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = db.Ping(); err != nil {
		panic(op + ":" + err.Error())
	}

	return &Repository{
		db: db,
	}, nil
}

func (s *Repository) Close() error {
	return s.db.Close()
}

func (s *Repository) SaveUser(ctx context.Context, username string, passHash []byte) (int64, error) {
	const op = "storage.postgres.SaveUser"

	var userID int64
	query := "INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id"

	err := s.db.QueryRowContext(ctx, query, username, passHash).Scan(&userID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (s *Repository) GetUser(ctx context.Context, username string) (*models.User, error) {
	const op = "storage.postgres.User"

	stmt, err := s.db.Prepare("SELECT id, username, password_hash FROM users WHERE username = $1")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, username)

	var user models.User
	err = row.Scan(&user.UID, &user.Username, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, ErrNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Repository) GetUsersDetails(ctx context.Context, userIDs []int64) ([]*models.User, error) {
	query := "SELECT id, username FROM users WHERE id = ANY($1)"

	var res []*models.User
	if err := s.db.SelectContext(ctx, &res, query, pq.Array(userIDs)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return res, nil
}
