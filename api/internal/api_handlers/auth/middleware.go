package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	authclient "github.com/liriquew/control_system/api/internal/grpc/clients/auth"
	"github.com/liriquew/control_system/api/pkg/logger/sl"
	"google.golang.org/grpc/metadata"
)

type UserID struct{}

type AuthMiddleware struct {
	log    *slog.Logger
	client authclient.AuthClient
}

func NewAuthMiddleware(log *slog.Logger, client authclient.AuthClient) *AuthMiddleware {
	return &AuthMiddleware{
		log:    log,
		client: client,
	}
}

func (a *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString, found := strings.CutPrefix(tokenString, "Bearer ")
		if !found {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}

		userID, err := a.client.Authenticate(r.Context(), tokenString)
		if err != nil {
			a.log.Error("error while authenticate user:", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), UserID{}, userID)

		md := metadata.Pairs("user-id", strconv.FormatInt(userID, 10))
		ctx = metadata.NewOutgoingContext(ctx, md)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) int64 {
	id, _ := r.Context().Value(UserID{}).(int64)
	return id
}
