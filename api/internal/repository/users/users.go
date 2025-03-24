package users_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/liriquew/control_system/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	ErrUserNotFound      = fmt.Errorf("пользователь не найден")
	ErrUserAlreadyExists = fmt.Errorf("пользователь уже существует")
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) (*UserRepository, error) {
	return &UserRepository{
		db: db,
	}, nil
}

func (s *UserRepository) CreateUser(ctx context.Context, username, password string) (int64, error) {
	user := &models.User{}

	query := "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id"

	err := s.db.QueryRowContext(
		ctx,
		query,
		username, password,
	).Scan(&user.ID)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return 0, ErrUserAlreadyExists
		}
		return 0, err
	}

	return user.ID, nil
}

func (s *UserRepository) GetUserByID(ctx context.Context, uid int64) (*models.User, error) {
	query := "SELECT * FROM users WHERE id = $1"

	var user models.User
	err := s.db.GetContext(ctx, &user, query, uid)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := "SELECT * FROM users WHERE username = $1"

	var user models.User
	err := s.db.GetContext(ctx, &user, query, username)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
