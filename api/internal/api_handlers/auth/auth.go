package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time_manage/internal/models"
	service "time_manage/internal/service/users"
)

const (
	headerContentType = "Content-Type"
	jsonContentType   = "application/json"
)

type AuthAPI interface {
	SignIn(w http.ResponseWriter, r *http.Request)
	SignUp(w http.ResponseWriter, r *http.Request)
	AuthRequired(next http.Handler) http.Handler
}

type userService interface {
	AuthenticateUser(ctx context.Context, user *models.User) (*models.User, error)
	RegisterUser(ctx context.Context, user *models.User) (*models.User, error)
}

type Auth struct {
	userService userService
	infoLog     *log.Logger
	errorLog    *log.Logger
}

func New(infoLog *log.Logger, errorLog *log.Logger, userService userService) *Auth {
	return &Auth{
		userService: userService,
		infoLog:     infoLog,
		errorLog:    errorLog,
	}
}

type tokenJWT struct {
	Token string `json:"token"`
}

func (api *Auth) SignIn(w http.ResponseWriter, r *http.Request) {
	user, err := models.UserModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "error while reading request body", http.StatusBadRequest)
		return
	}

	user, err = api.userService.AuthenticateUser(r.Context(), user)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "bad request body", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusBadRequest)
			return
		}

		api.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jwtToken, err := GenerateJWT(user.ID)
	if err != nil {
		api.errorLog.Println("JWT error:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(tokenJWT{jwtToken})
}

func (api *Auth) SignUp(w http.ResponseWriter, r *http.Request) {
	user, err := models.UserModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "error while reading request body", http.StatusBadRequest)
		return
	}

	user, err = api.userService.RegisterUser(r.Context(), user)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			http.Error(w, "bad request body", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrAlreadyExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}

		api.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jwtToken, err := GenerateJWT(user.ID)
	if err != nil {
		api.errorLog.Println("JWT error:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(tokenJWT{jwtToken})
}
