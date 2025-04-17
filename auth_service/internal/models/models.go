package models

import auth_pb "github.com/liriquew/control_system/services_protos/auth_service"

type User struct {
	UID      int64  `db:"id"`
	Username string `db:"username"`
	PassHash []byte
}

func ConvertUserToProto(user *User) *auth_pb.UserProfileDetails {
	return &auth_pb.UserProfileDetails{
		Username: user.Username,
		UserID:   user.UID,
	}
}

func ConvertUsersToProto(users []*User) []*auth_pb.UserProfileDetails {
	res := make([]*auth_pb.UserProfileDetails, 0, len(users))
	for _, user := range users {
		res = append(res, ConvertUserToProto(user))
	}
	return res
}
