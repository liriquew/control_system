package groups

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/liriquew/control_system/groups_service/internal/models"
	repository "github.com/liriquew/control_system/groups_service/internal/repository"
	"github.com/liriquew/control_system/groups_service/pkg/logger/sl"
	grpc_pb "github.com/liriquew/control_system/services_protos/groups_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Repository interface {
	CheckAdminPermission(ctx context.Context, userID, groupID int64) error
	CheckEditorPermission(ctx context.Context, userID, groupID int64) error
	CheckAccess(ctx context.Context, userID, groupID int64) error

	CreateGroup(ctx context.Context, group *grpc_pb.Group) (int64, error)
	ListUserGroups(ctx context.Context, userID int64, offset int64) ([]*models.Group, error)
	GetGroup(ctx context.Context, userID, groupID int64) (*models.Group, error)
	DeleteGroup(ctx context.Context, ownerID, groupID int64) error
	UpdateGroup(ctx context.Context, group *grpc_pb.Group) error
	ListGroupMembers(ctx context.Context, groupID int64) ([]*models.GroupMember, error)
	AddGroupMember(ctx context.Context, member *grpc_pb.GroupMember) error
	RemoveGroupMember(ctx context.Context, member *grpc_pb.GroupMember) error
	ChangeMemberRole(ctx context.Context, ownerID int64, member *grpc_pb.GroupMember) error
}

type Service struct {
	grpc_pb.UnimplementedGroupsServer
	repository Repository
	log        *slog.Logger
}

func NewServerAPI(log *slog.Logger, taskRepository Repository) *Service {
	return &Service{
		log:        log,
		repository: taskRepository,
	}
}

func (s *Service) Authenticate(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.log.Error("error while extracting metadata")
		return 0, status.Error(codes.Unauthenticated, "missing metadata")
	}

	AuthParams := md.Get("user-id")
	if len(AuthParams) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing user-idmetadata")
	}
	userID, err := strconv.ParseInt(AuthParams[0], 10, 64)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "invalid user-id metadata")
	}

	return userID, nil
}

func (s *Service) CreateGroup(ctx context.Context, group *grpc_pb.Group) (*grpc_pb.GroupID, error) {
	userID, err := s.Authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}
	group.OwnerID = userID

	groupID, err := s.repository.CreateGroup(ctx, group)
	if err != nil {
		if errors.Is(err, repository.ErrInvalideRole) {
			return nil, status.Error(codes.InvalidArgument, "invalide role param")
		}

		s.log.Error("error while creating group", sl.Err(err))
		return nil, err
	}

	return &grpc_pb.GroupID{
		ID: groupID,
	}, nil
}

func (s *Service) ListUserGroups(ctx context.Context, offset *grpc_pb.Offset) (*grpc_pb.GroupsList, error) {
	userID, err := s.Authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}

	groups, err := s.repository.ListUserGroups(ctx, userID, offset.Offset)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		s.log.Error("error while listing user's groups", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	resp := make([]*grpc_pb.Group, 0, len(groups))
	for _, g := range groups {
		resp = append(resp, models.ConvertGroupToProto(g))
	}

	s.log.Debug("grouos", slog.Any("groups", resp))

	return &grpc_pb.GroupsList{
		Groups: resp,
	}, nil
}

func (s *Service) GetGroup(ctx context.Context, groupID *grpc_pb.GroupID) (*grpc_pb.Group, error) {
	userID, err := s.Authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}
	if err := s.repository.CheckAccess(ctx, userID, groupID.ID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		s.log.Error("error while checking access", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	group, err := s.repository.GetGroup(ctx, userID, groupID.ID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "nothing found")
		}

		s.log.Error("error while getting group", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return models.ConvertGroupToProto(group), nil
}

func (s *Service) DeleteGroup(ctx context.Context, groupID *grpc_pb.GroupID) (*emptypb.Empty, error) {
	userID, err := s.Authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}
	if err := s.repository.CheckAdminPermission(ctx, userID, groupID.ID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		s.log.Error("error while checking admin permission", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	if err := s.repository.DeleteGroup(ctx, userID, groupID.ID); err != nil {
		s.log.Error("error while deleting group member", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) UpdateGroup(ctx context.Context, group *grpc_pb.Group) (*emptypb.Empty, error) {
	userID, err := s.Authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}

	if err := s.repository.CheckEditorPermission(ctx, userID, group.ID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		s.log.Error("error while checking editor permission", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}
	group.OwnerID = userID

	if err := s.repository.UpdateGroup(ctx, group); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return &emptypb.Empty{}, status.Error(codes.NotFound, "group not found")
		}

		s.log.Error("error while updating group", sl.Err(err))
		return &emptypb.Empty{}, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) ListGroupMembers(ctx context.Context, groupID *grpc_pb.GroupID) (*grpc_pb.GroupMembersList, error) {
	userID, err := s.Authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}
	if err := s.repository.CheckAccess(ctx, userID, groupID.ID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		s.log.Error("error while checking access", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	members, err := s.repository.ListGroupMembers(ctx, groupID.ID)
	if err != nil {
		s.log.Error("error while listing group members", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	resp := make([]*grpc_pb.GroupMember, 0, len(members))
	for _, gm := range members {
		resp = append(resp, models.ConvertMemberToProto(gm))
	}

	return &grpc_pb.GroupMembersList{
		Members: resp,
	}, nil
}

func (s *Service) AddGroupMember(ctx context.Context, GroupMember *grpc_pb.GroupMember) (*emptypb.Empty, error) {
	userID, err := s.Authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}
	if err := s.repository.CheckAdminPermission(ctx, userID, GroupMember.GroupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		s.log.Error("error while checking admin permission", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	if err := s.repository.AddGroupMember(ctx, GroupMember); err != nil {
		if errors.Is(err, repository.ErrNotExists) {
			return nil, status.Error(codes.NotFound, "group not found")
		}
		if errors.Is(err, repository.ErrAlreadyInGroup) {
			return nil, status.Error(codes.AlreadyExists, "user already in group")
		}
		if errors.Is(err, repository.ErrInvalideRole) {
			return nil, status.Error(codes.InvalidArgument, "bad role param")
		}

		s.log.Error("error while adding group member", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil

}

func (s *Service) RemoveGroupMember(ctx context.Context, GroupMember *grpc_pb.GroupMember) (*emptypb.Empty, error) {
	userID, err := s.Authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}
	if err := s.repository.CheckAdminPermission(ctx, userID, GroupMember.GroupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		s.log.Error("error while checking admin permission", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	if err := s.repository.RemoveGroupMember(ctx, GroupMember); err != nil {
		s.log.Error("error while removing group member", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) ChangeMemberRole(ctx context.Context, GroupMember *grpc_pb.GroupMember) (*emptypb.Empty, error) {
	userID, err := s.Authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}
	if err := s.repository.CheckAdminPermission(ctx, userID, GroupMember.GroupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		s.log.Error("error while checking access", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	if err := s.repository.ChangeMemberRole(ctx, userID, GroupMember); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "member not found")
		}
		if errors.Is(err, repository.ErrInvalideRole) {
			return nil, status.Error(codes.InvalidArgument, "bad role")
		}

		s.log.Error("error while updating group member", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) CheckAdminPermission(ctx context.Context, groupMember *grpc_pb.GroupMember) (*emptypb.Empty, error) {
	if err := s.repository.CheckAdminPermission(ctx, groupMember.UserID, groupMember.GroupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, status.Error(codes.NotFound, "user with permissions not found")
		}

		s.log.Error("error while checking user admin permission", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) CheckEditorPermission(ctx context.Context, groupMember *grpc_pb.GroupMember) (*emptypb.Empty, error) {
	if err := s.repository.CheckEditorPermission(ctx, groupMember.UserID, groupMember.GroupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, status.Error(codes.NotFound, "user with permissions not found")
		}

		s.log.Error("error while checking user editor permission", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) CheckMemberPermission(ctx context.Context, groupMember *grpc_pb.GroupMember) (*emptypb.Empty, error) {
	if err := s.repository.CheckAccess(ctx, groupMember.UserID, groupMember.GroupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, status.Error(codes.NotFound, "user with permissions not found")
		}

		s.log.Error("error while checking user access", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}
