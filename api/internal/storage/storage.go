package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time_manage/internal/config"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var (
	ErrUserNotFound      = fmt.Errorf("Пользователь не найден")
	ErrUserAlreadyExists = fmt.Errorf("Пользователь уже существует")
)

type User struct {
	ID       int64  `json:"uid" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type Storage struct {
	db *sqlx.DB
}

func New(cfg *config.StorageConfig) (*Storage, error) {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) CreateUser(ctx context.Context, username, password string) (int64, error) {
	user := &User{}

	query := "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id"

	err := s.db.QueryRowContext(
		ctx,
		query,
		username, password,
	).Scan(&user.ID)

	if err != nil {
		// TODO: check already exists
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return 0, ErrUserAlreadyExists
		}
		return 0, err
	}

	return user.ID, nil
}

func (s *Storage) GetUserByID(ctx context.Context, uid int64) (*User, error) {
	query := "SELECT * FROM users WHERE id = $1"

	var user User
	err := s.db.GetContext(ctx, &user, query, uid)

	if err != nil {
		// TODO: check error
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *Storage) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := "SELECT * FROM users WHERE username = $1"

	var user User
	err := s.db.GetContext(ctx, &user, query, username)
	fmt.Println(user, err)

	if err != nil {
		// TODO: check error
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
