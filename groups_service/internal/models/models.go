package models

import (
	"time"

	grpc_pb "github.com/liriquew/control_system/services_protos/groups_service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Group struct {
	ID          int64     `json:"id" db:"id"`
	OwnerID     int64     `json:"owner_id" db:"owner_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

func ConvertGroupToProto(grp *Group) *grpc_pb.Group {
	return &grpc_pb.Group{
		ID:          grp.ID,
		OwnerID:     grp.OwnerID,
		Name:        grp.Name,
		Description: grp.Description,
		CreatedAt:   timestamppb.New(grp.CreatedAt),
	}
}

type GroupMember struct {
	GroupID int64  `json:"group_id" db:"group_id"`
	UserID  int64  `json:"user_id" db:"user_id"`
	Role    string `json:"role" db:"role"`
}

func ConvertMemberToProto(gm *GroupMember) *grpc_pb.GroupMember {
	return &grpc_pb.GroupMember{
		UserID:  gm.UserID,
		GroupID: gm.GroupID,
		Role:    gm.Role,
	}
}
