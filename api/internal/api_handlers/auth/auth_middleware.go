package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type UIDInterface struct{}

var JWTsecretKey = []byte("anyEps")

func (api *Auth) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, err := CheckJWT(r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UIDInterface{}, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CheckJWT(r *http.Request) (int64, error) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return 0, fmt.Errorf("missing Authorization header")
	}

	tokenString, found := strings.CutPrefix(tokenString, "Bearer ")

	if !found {
		return 0, fmt.Errorf("missing Authorization header")
	}

	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JWTsecretKey, nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	uidFloat, ok := (*claims)["uid"].(float64)
	if !ok {
		return 0, fmt.Errorf("uid not found or invalid type in JWT token")
	}
	uid := int64(uidFloat)
	return int64(uid), nil
}

func GenerateJWT(uid int64) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 12).Unix(),
		"iat": time.Now().Unix(),
		"uid": uid,
	}

	tokenString, err := token.SignedString(JWTsecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
