package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time_manage/internal/storage"
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

type Auth struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	storage  *storage.Storage
}

func New(infoLog *log.Logger, errorLog *log.Logger, storage *storage.Storage) *Auth {
	return &Auth{
		infoLog:  infoLog,
		errorLog: errorLog,
		storage:  storage,
	}
}

type tokenJWT struct {
	Token string `json:"token"`
}

func (api *Auth) SignIn(w http.ResponseWriter, r *http.Request) {
	// TODO: implement sighIn
	api.infoLog.Println("Sing In")

	var user storage.User
	err := json.NewDecoder(r.Body).Decode(&user)
	api.infoLog.Println("User:", user)
	if err != nil || user.Username == "" || user.Password == "" {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}

	userFromStorage, err := api.storage.GetUserByUsername(r.Context(), user.Username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			http.Error(w, "Пользователь не найден", http.StatusBadRequest)
			return
		}
		api.errorLog.Println("Storage error:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.Password != userFromStorage.Password {
		http.Error(w, "Неверный логин / пароль", http.StatusUnauthorized)
		return
	}

	jwtToken, err := GenerateJWT(userFromStorage.ID)
	if err != nil {
		api.errorLog.Println("JWT error:", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(tokenJWT{jwtToken})
}

func (api *Auth) SignUp(w http.ResponseWriter, r *http.Request) {
	// TODO: implement signUp
	api.infoLog.Println("sign Up")

	var user storage.User
	err := json.NewDecoder(r.Body).Decode(&user)
	api.infoLog.Println("User:", user)
	if err != nil || user.Username == "" || user.Password == "" {
		http.Error(w, "Ошибка чтения тела запроса", http.StatusBadRequest)
		return
	}

	uid, err := api.storage.CreateUser(r.Context(), user.Username, user.Password)
	if err != nil {
		api.errorLog.Println("Storage error:", err.Error())
		if errors.Is(storage.ErrUserAlreadyExists, err) {
			http.Error(w, "Пользователь с таким именем уже существует", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jwtToken, err := GenerateJWT(uid)
	if err != nil {
		api.errorLog.Println("JWT error:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set(headerContentType, jsonContentType)
	json.NewEncoder(w).Encode(tokenJWT{jwtToken})
}
