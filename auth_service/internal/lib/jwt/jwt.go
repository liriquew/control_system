package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var Secret string

// NewToken creates new JWT token for given user and app.
func NewToken(UID int64) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = UID

	tokenString, err := token.SignedString([]byte(Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func Validate(tokenString string) (int64, error) {
	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(Secret), nil
	})

	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	return int64((*claims)["uid"].(float64)), nil // да
}
