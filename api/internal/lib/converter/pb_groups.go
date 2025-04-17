package converter

import (
	"github.com/liriquew/control_system/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"

	grps_pb "github.com/liriquew/control_system/services_protos/groups_service"
)

func ConvertGroupToProto(group *models.Group) *grps_pb.Group {
	return &grps_pb.Group{
		ID:          group.ID,
		OwnerID:     group.OwnerID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   timestamppb.New(group.CreatedAt),
	}
}

func ConvertGroupToModel(group *grps_pb.Group) *models.Group {
	return &models.Group{
		ID:          group.ID,
		OwnerID:     group.OwnerID,
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt.AsTime(),
	}
}

func ConvertGroupsToModel(groups []*grps_pb.Group) []*models.Group {
	res := make([]*models.Group, 0, len(groups))
	for _, group := range groups {
		res = append(res, ConvertGroupToModel(group))
	}
	return res
}

func ConvertGroupMemberToProto(gm *models.GroupMember) *grps_pb.GroupMember {
	return &grps_pb.GroupMember{
		GroupID: gm.GroupID,
		UserID:  gm.UserID,
		Role:    gm.Role,
	}
}

func ConvertGroupMemberToModel(gm *grps_pb.GroupMember) *models.GroupMember {
	return &models.GroupMember{
		GroupID: gm.GroupID,
		UserID:  gm.UserID,
		Role:    gm.Role,
	}
}

func ConvertGroupMembersToModel(gms []*grps_pb.GroupMember) []*models.GroupMember {
	res := make([]*models.GroupMember, 0, len(gms))
	for _, gm := range gms {
		res = append(res, ConvertGroupMemberToModel(gm))
	}
	return res
}
