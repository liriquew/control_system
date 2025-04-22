package groupsclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"

	"github.com/liriquew/control_system/internal/lib/config"
	"github.com/liriquew/control_system/internal/lib/converter"
	"github.com/liriquew/control_system/internal/models"
	grps_pb "github.com/liriquew/control_system/services_protos/groups_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GRPCGroupsClient struct {
	client grps_pb.GroupsClient
	log    *slog.Logger
}

func NewGroupsClient(log *slog.Logger, cfg config.ClientConfig) (*GRPCGroupsClient, error) {
	const op = "authclient.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(cfg.Retries)),
		grpcretry.WithPerRetryTimeout(cfg.Timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	address := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	cc, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &GRPCGroupsClient{
		client: grps_pb.NewGroupsClient(cc),
		log:    log,
	}, nil
}

func InterceptorLogger(log *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		log.Log(ctx, slog.Level(level), msg, fields...)
	})
}

var (
	ErrNotUserFound     = errors.New("")
	ErrBadRoleParam     = errors.New("")
	ErrNotFound         = errors.New("not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrNothingToUpdate  = errors.New("nothing to update, empty fields")
	ErrAlreadyExists    = errors.New("already exists")
)

type GroupsClient interface {
	CreateGroup(ctx context.Context, group *models.Group) (int64, error)
	ListUserGroups(ctx context.Context, offset int64) ([]*models.Group, error)
	GetGroup(ctx context.Context, groupID int64) (*models.Group, error)
	DeleteGroup(ctx context.Context, groupID int64) error
	UpdateGroup(ctx context.Context, group *models.Group) error
	ListGroupMembers(ctx context.Context, groupID int64) ([]*models.GroupMember, error)
	AddGroupMember(ctx context.Context, groupMember *models.GroupMember) error
	RemoveGroupMember(ctx context.Context, groupMember *models.GroupMember) error
	ChangeMemberRole(ctx context.Context, groupMember *models.GroupMember) error
	CheckAdminPermission(ctx context.Context, userID, groupID int64) error
	CheckEditorPermission(ctx context.Context, userID, groupID int64) error
	CheckMemberPermission(ctx context.Context, userID, groupID int64) error
}

func (g *GRPCGroupsClient) CreateGroup(ctx context.Context, group *models.Group) (int64, error) {
	resp, err := g.client.CreateGroup(ctx, converter.ConvertGroupToProto(group))

	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return 0, fmt.Errorf("%w%s", ErrNotUserFound, st.Message())
			case codes.InvalidArgument:
				return 0, fmt.Errorf("%w%s", ErrBadRoleParam, st.Message())
			}
		}

		return 0, err
	}

	return resp.ID, nil
}

func (g *GRPCGroupsClient) ListUserGroups(ctx context.Context, offset int64) ([]*models.Group, error) {
	g.log.Debug("this is offset", slog.Int64("offset", offset))
	resp, err := g.client.ListUserGroups(ctx, &grps_pb.Offset{
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	return converter.ConvertGroupsToModel(resp.Groups), nil
}

func (g *GRPCGroupsClient) GetGroup(ctx context.Context, groupID int64) (*models.Group, error) {
	resp, err := g.client.GetGroup(ctx, &grps_pb.GroupID{
		ID: groupID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, ErrNotFound
			case codes.PermissionDenied:
				return nil, ErrPermissionDenied
			}
		}

		return nil, err
	}

	return converter.ConvertGroupToModel(resp), nil
}

func (g *GRPCGroupsClient) DeleteGroup(ctx context.Context, groupID int64) error {
	_, err := g.client.DeleteGroup(ctx, &grps_pb.GroupID{
		ID: groupID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.PermissionDenied:
				return ErrPermissionDenied
			}
		}

		return err
	}

	return nil
}

func (g *GRPCGroupsClient) UpdateGroup(ctx context.Context, group *models.Group) error {
	_, err := g.client.UpdateGroup(ctx, converter.ConvertGroupToProto(group))
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return ErrNotFound
			case codes.PermissionDenied:
				return ErrPermissionDenied
			}
		}

		return err
	}

	return nil
}

func (g *GRPCGroupsClient) ListGroupMembers(ctx context.Context, groupID int64) ([]*models.GroupMember, error) {
	resp, err := g.client.ListGroupMembers(ctx, &grps_pb.GroupID{
		ID: groupID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.PermissionDenied:
				return nil, ErrPermissionDenied
			}
		}

		return nil, err
	}

	return converter.ConvertGroupMembersToModel(resp.Members), nil
}

func (g *GRPCGroupsClient) AddGroupMember(ctx context.Context, groupMember *models.GroupMember) error {
	g.log.Debug("group member in client", slog.Any("gm", groupMember))
	_, err := g.client.AddGroupMember(ctx, converter.ConvertGroupMemberToProto(groupMember))
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.PermissionDenied:
				return ErrPermissionDenied
			case codes.NotFound:
				return ErrNotFound
			case codes.AlreadyExists:
				return ErrAlreadyExists
			case codes.InvalidArgument:
				return ErrBadRoleParam
			}
		}

		return err
	}

	return nil
}

func (g *GRPCGroupsClient) RemoveGroupMember(ctx context.Context, groupMember *models.GroupMember) error {
	_, err := g.client.RemoveGroupMember(ctx, converter.ConvertGroupMemberToProto(groupMember))
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.PermissionDenied:
				return ErrPermissionDenied
			}
		}
	}

	return nil
}

func (g *GRPCGroupsClient) ChangeMemberRole(ctx context.Context, groupMember *models.GroupMember) error {
	_, err := g.client.ChangeMemberRole(ctx, converter.ConvertGroupMemberToProto(groupMember))
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return ErrBadRoleParam
			case codes.PermissionDenied:
				return ErrPermissionDenied
			case codes.NotFound:
				return ErrNotFound
			}
		}

		return err
	}

	return nil
}

func (g *GRPCGroupsClient) CheckAdminPermission(ctx context.Context, userID, groupID int64) error {
	g.log.Debug("admin permission credentials", slog.Int64("userID", userID), slog.Int64("groupID", groupID))
	_, err := g.client.CheckAdminPermission(ctx, &grps_pb.GroupMember{
		UserID:  userID,
		GroupID: groupID,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return ErrPermissionDenied
		}

		return err
	}

	return nil
}

func (g *GRPCGroupsClient) CheckEditorPermission(ctx context.Context, userID, groupID int64) error {
	_, err := g.client.CheckEditorPermission(ctx, &grps_pb.GroupMember{
		UserID:  userID,
		GroupID: groupID,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return ErrPermissionDenied
		}

		return err
	}

	return nil
}

func (g *GRPCGroupsClient) CheckMemberPermission(ctx context.Context, userID, groupID int64) error {
	_, err := g.client.CheckMemberPermission(ctx, &grps_pb.GroupMember{
		UserID:  userID,
		GroupID: groupID,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return ErrPermissionDenied
		}

		return err
	}

	return nil
}
