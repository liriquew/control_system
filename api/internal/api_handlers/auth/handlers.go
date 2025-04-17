package auth

import (
	"errors"
	"log/slog"
	"net/http"

	authclient "github.com/liriquew/control_system/internal/grpc/clients/auth"
	jsontools "github.com/liriquew/control_system/internal/lib/json_tools"
	"github.com/liriquew/control_system/internal/models"
	"github.com/liriquew/control_system/pkg/logger/sl"
)

type AuthAPI interface {
	SignIn(w http.ResponseWriter, r *http.Request)
	SignUp(w http.ResponseWriter, r *http.Request)
}

type Auth struct {
	authClient authclient.AuthClient
	log        *slog.Logger
}

func NewAuthService(log *slog.Logger, authClient authclient.AuthClient) *Auth {
	return &Auth{
		authClient: authClient,
		log:        log,
	}
}

type tokenJWT struct {
	Token string `json:"token"`
}

func (a *Auth) SignIn(w http.ResponseWriter, r *http.Request) {
	userCreds, err := models.UserModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "error while reading request body", http.StatusBadRequest)
		return
	}

	if userCreds.Username == "" || userCreds.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jwtToken, err := a.authClient.SignIn(r.Context(), userCreds)
	if err != nil {
		if errors.Is(err, authclient.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, authclient.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}

		a.log.Error("error wile sign in user", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WtiteJSON(w, tokenJWT{jwtToken})
}

func (api *Auth) SignUp(w http.ResponseWriter, r *http.Request) {
	userCreds, err := models.UserModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "error while reading request body", http.StatusBadRequest)
		return
	}

	if userCreds.Username == "" || userCreds.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jwtToken, err := api.authClient.SignUp(r.Context(), userCreds)
	if err != nil {
		if errors.Is(err, authclient.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusBadRequest)
			return
		}
		if errors.Is(err, authclient.ErrAlreadyExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}

		api.log.Error("error while sign up user", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WtiteJSON(w, tokenJWT{jwtToken})
}
