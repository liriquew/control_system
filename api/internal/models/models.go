package models

import (
	"encoding/json"
	"io"

	auth_pb "github.com/liriquew/control_system/services_protos/auth_service"
)

type User struct {
	ID       int64  `json:"uid" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type UsersDetails struct {
	Username string `json:"username"`
}

func Details(user *User) *UsersDetails {
	return &UsersDetails{
		Username: user.Username,
	}
}

func UserModelFromJson(jsonBody io.ReadCloser) (*User, error) {
	var user User
	err := json.NewDecoder(jsonBody).Decode(&user)

	return &user, err
}

func ConvertUserFromProto(user *auth_pb.UserProfileDetails) *User {
	return &User{
		ID:       user.UserID,
		Username: user.Username,
	}
}

func ConvertUsersFromProto(users []*auth_pb.UserProfileDetails) []*User {
	res := make([]*User, 0, len(users))
	for _, user := range users {
		res = append(res, ConvertUserFromProto(user))
	}
	return res
}
