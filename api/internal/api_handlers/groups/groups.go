package groups

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/liriquew/control_system/internal/api_handlers/auth"
	"github.com/liriquew/control_system/internal/entities"
	jsontools "github.com/liriquew/control_system/internal/lib/json_tools"
	"github.com/liriquew/control_system/internal/models"
	service "github.com/liriquew/control_system/internal/service/groups"
)

type GroupsAPI interface {
	CreateGroup(w http.ResponseWriter, r *http.Request)
	ListUserGroups(w http.ResponseWriter, r *http.Request)
	GetGroup(w http.ResponseWriter, r *http.Request)
	DeleteGroup(w http.ResponseWriter, r *http.Request)
	UpdateGroup(w http.ResponseWriter, r *http.Request)
	ListGroupMembers(w http.ResponseWriter, r *http.Request)
	AddGroupMember(w http.ResponseWriter, r *http.Request)
	RemoveGroupMember(w http.ResponseWriter, r *http.Request)
	ChangeMemberRole(w http.ResponseWriter, r *http.Request)
	CreateGroupGraph(w http.ResponseWriter, r *http.Request)
	ListGroupGraphs(w http.ResponseWriter, r *http.Request)
}

type GroupServiceInterface interface {
	CreateGroup(ctx context.Context, group *models.Group) (*models.Group, error)
	ListUserGroups(ctx context.Context, userID int64) ([]*models.Group, error)
	GetGroup(ctx context.Context, userID, groupID int64) (*models.Group, error)
	DeleteGroup(ctx context.Context, ownerID, groupID int64) error
	UpdateGroup(ctx context.Context, group *models.Group) error
	ListGroupMembers(ctx context.Context, userID, groupID int64) ([]*models.User, error)
	AddGroupMember(ctx context.Context, ownerID int64, member *models.GroupMember) (*models.GroupMember, error)
	RemoveGroupMember(ctx context.Context, ownerID int64, member *models.GroupMember) error
	ChangeMemberRole(ctx context.Context, ownerID int64, member *models.GroupMember) error
	CreateGroupGraph(ctx context.Context, graph *entities.GraphWithNodes) (*models.Graph, error)
	ListGroupGraphs(ctx context.Context, userID, groupID int64) ([]*entities.GraphWithNodes, error)
}

type Group struct {
	groupService GroupServiceInterface
	infoLog      *log.Logger
	errorLog     *log.Logger
}

func New(groupService GroupServiceInterface, infoLog, errorLog *log.Logger) *Group {
	return &Group{
		groupService: groupService,
		infoLog:      infoLog,
		errorLog:     errorLog,
	}
}

func (g *Group) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)

	group, err := models.GroupModelFromJson(r.Body)
	if err != nil {
		g.errorLog.Println(err)
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	group.OwnerID = userID

	group, err = g.groupService.CreateGroup(r.Context(), group)
	if err != nil {
		if errors.Is(err, service.ErrWrongOwnerID) {
			http.Error(w, "bad owner id", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}

func (g *Group) GetGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	groupID := r.Context().Value(GroupID{}).(int64)

	group, err := g.groupService.GetGroup(r.Context(), userID, groupID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}

func (g *Group) ListUserGroups(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)

	groups, err := g.groupService.ListUserGroups(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(groups)
}

func (g *Group) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	groupID := r.Context().Value(GroupID{}).(int64)

	err := g.groupService.DeleteGroup(r.Context(), userID, groupID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Group) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	groupID := r.Context().Value(GroupID{}).(int64)

	group, err := models.GroupModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	group.ID = groupID
	group.OwnerID = userID

	err = g.groupService.UpdateGroup(r.Context(), group)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrNothingToUpdate) {
			http.Error(w, "nothing to update (empty updatable fields)", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Group) ListGroupMembers(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	groupID := r.Context().Value(GroupID{}).(int64)

	members, err := g.groupService.ListGroupMembers(r.Context(), userID, groupID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(members)
}

func (g *Group) AddGroupMember(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	groupID := r.Context().Value(GroupID{}).(int64)

	groupMember, err := models.GroupMemberModelFromJson(r.Body)
	if err != nil {
		g.errorLog.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	groupMember.GroupID = groupID

	// TODO user with ID = userID in this case can add users only if he is admin btw
	// check it in service layer
	members, err := g.groupService.AddGroupMember(r.Context(), userID, groupMember)
	if err != nil {
		if errors.Is(err, service.ErrNotExists) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrInvalideRole) {
			http.Error(w, "bad role", http.StatusBadRequest)
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(members)
}

func (g *Group) RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	groupID := r.Context().Value(GroupID{}).(int64)
	memberID := r.Context().Value(MemberID{}).(int64)

	groupMember := &models.GroupMember{
		UserID:  memberID,
		GroupID: groupID,
	}

	// TODO user with ID = userID in this case can add users only if he is admin btw
	// check it in service layer
	err := g.groupService.RemoveGroupMember(r.Context(), userID, groupMember)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Group) ChangeMemberRole(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	groupID := r.Context().Value(GroupID{}).(int64)
	memberID := r.Context().Value(MemberID{}).(int64)
	newRole := r.URL.Query().Get("newRole")

	groupMember := &models.GroupMember{
		GroupID: groupID,
		UserID:  memberID,
		Role:    newRole,
	}

	// TODO user with ID = userID in this case can add users only if he is admin btw
	// check it in service layer
	err := g.groupService.ChangeMemberRole(r.Context(), userID, groupMember)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrInvalideRole) {
			http.Error(w, "bad role param", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Group) CreateGroupGraph(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	groupID := r.Context().Value(GroupID{}).(int64)

	graph, err := entities.GraphWithNodesFromJSON(r.Body)
	if err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	graph.GraphInfo.GroupID = groupID
	graph.GraphInfo.CreatedBy = userID

	createdGraph, err := g.groupService.CreateGroupGraph(r.Context(), graph)
	if err != nil {
		if errors.Is(err, service.ErrNoNodes) {
			http.Error(w, "no nodes", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrCycleInGraph) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		g.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createdGraph)
}

func (g *Group) ListGroupGraphs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	groupID := r.Context().Value(GroupID{}).(int64)

	graphs, err := g.groupService.ListGroupGraphs(r.Context(), userID, groupID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.errorLog.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	jsontools.WtiteJSON(w, graphs)
}
