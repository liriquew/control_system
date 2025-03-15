package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time_manage/internal/entities"
	"time_manage/internal/models"
	repository "time_manage/internal/repository/groups"
	service "time_manage/internal/service/graphs"
)

type GroupsRepository interface {
	CheckAdminPermission(ctx context.Context, userID, groupID int64) error
	CheckEditorPermission(ctx context.Context, userID, groupID int64) error
	CheckAccess(ctx context.Context, userID, groupID int64) error

	CreateGroup(ctx context.Context, group *models.Group) (*models.Group, error)
	ListUserGroups(ctx context.Context, userID int64) ([]*models.Group, error)
	GetGroup(ctx context.Context, userID, groupID int64) (*models.Group, error)
	DeleteGroup(ctx context.Context, ownerID, groupID int64) (bool, error)
	UpdateGroup(ctx context.Context, group *models.Group) error
	ListGroupMembers(ctx context.Context, userID, groupID int64) ([]*models.User, error)
	AddGroupMember(ctx context.Context, ownerID int64, member *models.GroupMember) (*models.GroupMember, error)
	RemoveGroupMember(ctx context.Context, ownerID int64, member *models.GroupMember) (bool, error)
	ChangeMemberRole(ctx context.Context, ownerID int64, member *models.GroupMember) error
	Begin(ctx context.Context) (*sql.Tx, error)
	CreateGroupGraph(ctx context.Context, graph *models.Graph, nodes []*models.Node) (*models.Graph, error)
	ListGroupGraphs(ctx context.Context, userID int64, groupID int64) ([]*entities.GraphWithNodes, error)
}

type GroupService struct {
	groupRepo GroupsRepository
	infoLog   *log.Logger
	errorLog  *log.Logger
}

func NewGroupService(groupRepo GroupsRepository, infoLog, errorLog *log.Logger) (*GroupService, error) {
	return &GroupService{
		groupRepo: groupRepo,
		infoLog:   infoLog,
		errorLog:  errorLog,
	}, nil
}

var (
	ErrDenied = errors.New("deny")

	ErrNotFound        = errors.New("not found")
	ErrNotExists       = errors.New("not exists")
	ErrNothingToUpdate = errors.New("nothing to update (empty updatable fields)")

	ErrBadGroupParams = errors.New("bad params")
	ErrWrongOwnerID   = errors.New("bad owner id ")
	ErrMemberNotFound = errors.New("group not found")
	ErrNoNodes        = errors.New("empty nodes list")
	ErrInvalideRole   = errors.New("invalide role")
)

func (gs *GroupService) CreateGroup(ctx context.Context, group *models.Group) (*models.Group, error) {
	if !group.Validate() {
		return nil, ErrBadGroupParams
	}

	group, err := gs.groupRepo.CreateGroup(ctx, group)
	if err != nil {
		gs.errorLog.Println(err)
		if errors.Is(err, repository.ErrWrongOwnerID) {
			return nil, fmt.Errorf("%s%w", err.Error(), ErrWrongOwnerID)
		}

		return nil, err
	}

	return group, nil
}

func (gs *GroupService) GetGroup(ctx context.Context, userID, groupID int64) (*models.Group, error) {
	group, err := gs.groupRepo.GetGroup(ctx, userID, groupID)
	if err != nil {
		gs.errorLog.Println(err)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return group, err
}

func (gs *GroupService) ListUserGroups(ctx context.Context, userID int64) ([]*models.Group, error) {
	groups, err := gs.groupRepo.ListUserGroups(ctx, userID)
	if err != nil {
		gs.errorLog.Println(err)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return groups, nil
}

func (gs *GroupService) DeleteGroup(ctx context.Context, ownerID, groupID int64) error {
	if err := gs.groupRepo.CheckAdminPermission(ctx, ownerID, groupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return ErrDenied
		}
		return err
	}

	deleted, err := gs.groupRepo.DeleteGroup(ctx, ownerID, groupID)
	if err != nil && !deleted {
		gs.errorLog.Println(err)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}

		return err
	}

	return nil
}

func (gs *GroupService) UpdateGroup(ctx context.Context, group *models.Group) error {
	if err := gs.groupRepo.CheckAdminPermission(ctx, group.OwnerID, group.ID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return ErrDenied
		}
		return err
	}

	err := gs.groupRepo.UpdateGroup(ctx, group)
	if err != nil {
		gs.errorLog.Println(err)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		if errors.Is(err, repository.ErrNothingToUpdate) {
			return ErrNothingToUpdate
		}

		return err
	}

	return nil
}

func (gs *GroupService) ListGroupMembers(ctx context.Context, userID, groupID int64) ([]*models.User, error) {
	if err := gs.groupRepo.CheckAccess(ctx, userID, groupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, ErrDenied
		}
		return nil, err
	}

	users, err := gs.groupRepo.ListGroupMembers(ctx, userID, groupID)
	if err != nil {
		gs.errorLog.Println(err)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return users, err
}

func (gs *GroupService) AddGroupMember(ctx context.Context, ownerID int64, member *models.GroupMember) (*models.GroupMember, error) {
	if err := gs.groupRepo.CheckAdminPermission(ctx, ownerID, member.GroupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, ErrDenied
		}
		return nil, err
	}

	member, err := gs.groupRepo.AddGroupMember(ctx, ownerID, member)
	if err != nil {
		gs.errorLog.Println(err)
		if errors.Is(err, repository.ErrNotExists) {
			return nil, fmt.Errorf("%s%w", err.Error(), ErrNotExists)
		}
		if errors.Is(err, repository.ErrInvalideRole) {
			return nil, ErrInvalideRole
		}

		return nil, fmt.Errorf("AddGroupMember error: %w", err)
	}

	return member, nil
}

func (gs *GroupService) RemoveGroupMember(ctx context.Context, ownerID int64, member *models.GroupMember) error {
	if err := gs.groupRepo.CheckAdminPermission(ctx, ownerID, member.GroupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return ErrDenied
		}
		return err
	}

	removed, err := gs.groupRepo.RemoveGroupMember(ctx, ownerID, member)
	if err != nil && !removed {
		gs.errorLog.Println(err)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}

		return fmt.Errorf("RemoveGroupMember error: %w", err)
	}

	return nil
}

func (gs *GroupService) ChangeMemberRole(ctx context.Context, ownerID int64, member *models.GroupMember) error {
	if err := gs.groupRepo.CheckAdminPermission(ctx, ownerID, member.GroupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return ErrDenied
		}
		return err
	}

	err := gs.groupRepo.ChangeMemberRole(ctx, ownerID, member)
	if err != nil {
		gs.errorLog.Println(err)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		if errors.Is(err, repository.ErrInvalideRole) {
			return ErrInvalideRole
		}

		return err
	}

	return nil
}

func (gs *GroupService) CreateGroupGraph(ctx context.Context, graph *entities.GraphWithNodes) (*models.Graph, error) {
	if err := gs.groupRepo.CheckAdminPermission(ctx, graph.GraphInfo.CreatedBy, graph.GraphInfo.GroupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, ErrDenied
		}
		return nil, err
	}

	if len(graph.Nodes) == 0 {
		return nil, ErrNoNodes
	}

	if err := gs.groupRepo.CheckAdminPermission(ctx, graph.GraphInfo.CreatedBy, graph.GraphInfo.GroupID); err != nil {
		gs.errorLog.Println(err)
		if errors.Is(err, repository.ErrDenied) {
			return nil, ErrDenied
		}

		return nil, err
	}

	// check cycles in graph by topo sort algo (write this btw)

	graphModel, err := gs.groupRepo.CreateGroupGraph(ctx, &graph.GraphInfo, graph.Nodes)
	if err != nil {
		gs.errorLog.Println(err)
		return nil, err
	}

	return graphModel, nil
}

func (gs *GroupService) ListGroupGraphs(ctx context.Context, userID, groupID int64) ([]*entities.GraphWithNodes, error) {
	if err := gs.groupRepo.CheckAccess(ctx, userID, groupID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, ErrDenied
		}
		return nil, err
	}

	graphs, err := gs.groupRepo.ListGroupGraphs(ctx, userID, groupID)
	if err != nil {
		gs.errorLog.Println(err)
		if errors.Is(err, service.ErrNotFound) {
			return nil, ErrNotFound
		}

		return nil, err
	}
	return graphs, nil
}
