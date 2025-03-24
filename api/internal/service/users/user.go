package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/liriquew/control_system/internal/models"
	repository "github.com/liriquew/control_system/internal/repository/users"
)

type UserRepository interface {
	CreateUser(ctx context.Context, username, password string) (int64, error)
	GetUserByID(ctx context.Context, uid int64) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type UserService struct {
	userRepo UserRepository
	infolog  *log.Logger
	errorLog *log.Logger
}

func NewUserService(userRepo UserRepository, infoLog *log.Logger, errorLog *log.Logger) (*UserService, error) {
	return &UserService{
		userRepo: userRepo,
		infolog:  infoLog,
		errorLog: errorLog,
	}, nil
}

var (
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
	ErrUserNotFound       = fmt.Errorf("user not found")
	ErrAlreadyExists      = fmt.Errorf("user already exists")
)

func (us *UserService) AuthenticateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if user.Username == "" || user.Password == "" {
		return nil, ErrInvalidCredentials
	}

	userFromStorage, err := us.userRepo.GetUserByUsername(ctx, user.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		us.errorLog.Println("Storage error:", err)
		return nil, err
	}

	if user.Password != userFromStorage.Password {
		return nil, ErrInvalidCredentials
	}

	user.ID = userFromStorage.ID
	return user, nil
}

func (us *UserService) RegisterUser(ctx context.Context, user *models.User) (*models.User, error) {
	if user.Username == "" || user.Password == "" {
		return nil, ErrInvalidCredentials
	}

	uid, err := us.userRepo.CreateUser(ctx, user.Username, user.Password)
	if err != nil {
		us.errorLog.Println("Storage error:", err.Error())
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}

	user.ID = uid
	return user, nil
}
