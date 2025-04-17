package groups

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/liriquew/control_system/internal/api_handlers/auth"
	"github.com/liriquew/control_system/internal/entities"
	authclient "github.com/liriquew/control_system/internal/grpc/clients/auth"
	groupsclient "github.com/liriquew/control_system/internal/grpc/clients/groups"
	tasksclient "github.com/liriquew/control_system/internal/grpc/clients/tasks"
	jsontools "github.com/liriquew/control_system/internal/lib/json_tools"
	"github.com/liriquew/control_system/internal/models"
	"github.com/liriquew/control_system/pkg/logger/sl"
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
}

type Group struct {
	groupsClient groupsclient.GroupsClient
	tasksClient  tasksclient.TasksClient
	authClient   authclient.AuthClient
	log          *slog.Logger
}

func NewGroupsService(log *slog.Logger, groupsClient groupsclient.GroupsClient, tasksclient tasksclient.TasksClient, authClient authclient.AuthClient) *Group {
	return &Group{
		groupsClient: groupsClient,
		tasksClient:  tasksclient,
		authClient:   authClient,
		log:          log,
	}
}

func (g *Group) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserID{}).(int64)

	group, err := models.GroupModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if group.Name == "" {
		http.Error(w, "invalid name", http.StatusBadRequest)
		return
	}
	if group.Description == "" {
		http.Error(w, "invalid description", http.StatusBadRequest)
		return
	}

	group.OwnerID = userID

	groupID, err := g.groupsClient.CreateGroup(r.Context(), group)
	if err != nil {
		if errors.Is(err, groupsclient.ErrBadRoleParam) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		g.log.Error("error while creating group", slog.Any("group", group), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WriteInt64ID(w, groupID)
}

func (g *Group) ListUserGroups(w http.ResponseWriter, r *http.Request) {
	padding := GetPadding(r)

	groups, err := g.groupsClient.ListUserGroups(r.Context(), padding)
	if err != nil {
		g.log.Error("error while listing user's groups", slog.Int64("userID", padding), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WtiteJSON(w, groups)
}

func (g *Group) GetGroup(w http.ResponseWriter, r *http.Request) {
	groupID := GetGroupID(r)

	group, err := g.groupsClient.GetGroup(r.Context(), groupID)
	if err != nil {
		if errors.Is(err, groupsclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, groupsclient.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.log.Error("error while getting user group", slog.Int64("groupID", groupID), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tasks, err := g.tasksClient.GetGroupTasks(r.Context(), groupID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	g.log.Debug("group info in GetGroup", slog.Any("group info", entities.GroupWithTasks{
		Group: group,
		Tasks: tasks,
	}))

	jsontools.WtiteJSON(w, entities.GroupWithTasks{
		Group: group,
		Tasks: tasks,
	})
}

func (g *Group) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	groupID := GetGroupID(r)

	err := g.groupsClient.DeleteGroup(r.Context(), groupID)
	if err != nil {
		if errors.Is(err, groupsclient.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.log.Error("error while deleting group", slog.Int64("groupID", groupID), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Group) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	groupID := GetGroupID(r)

	group, err := models.GroupModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if group.Description == "" && group.Name == "" {
		http.Error(w, "empty unpdatable fields", http.StatusBadRequest)
		return
	}
	group.ID = groupID

	err = g.groupsClient.UpdateGroup(r.Context(), group)
	if err != nil {
		if errors.Is(err, groupsclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, groupsclient.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.log.Error("error while updating group", slog.Any("group", group), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Group) ListGroupMembers(w http.ResponseWriter, r *http.Request) {
	groupID := GetGroupID(r)

	members, err := g.groupsClient.ListGroupMembers(r.Context(), groupID)
	if err != nil {
		if errors.Is(err, groupsclient.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.log.Error("error while listing group memebers", slog.Int64("groupID", groupID), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userIDs := make([]int64, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
	}

	usersDetails, err := g.authClient.GetUsersDetails(r.Context(), userIDs)
	if err != nil {
		g.log.Error("error while getting users details", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	usersMap := make(map[int64]*models.UsersDetails, len(usersDetails))
	for _, user := range usersDetails {
		usersMap[user.ID] = models.Details(user)
	}

	resp := make([]*entities.GroupMemberWithDetails, 0, len(usersDetails))

	for _, member := range members {
		resp = append(resp, &entities.GroupMemberWithDetails{
			Member:  member,
			Details: usersMap[member.UserID],
		})
	}

	jsontools.WtiteJSON(w, resp)
}

func (g *Group) AddGroupMember(w http.ResponseWriter, r *http.Request) {
	groupID := GetGroupID(r)

	groupMember, err := models.GroupMemberModelFromJson(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	groupMember.GroupID = groupID

	err = g.groupsClient.AddGroupMember(r.Context(), groupMember)
	if err != nil {
		if errors.Is(err, groupsclient.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
		}
		if errors.Is(err, groupsclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, groupsclient.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if errors.Is(err, groupsclient.ErrBadRoleParam) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		g.log.Error("error while adding group member", slog.Any("group member", groupMember), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Group) RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	groupID := GetGroupID(r)
	memberID := GetGroupMemberID(r)

	groupMember := &models.GroupMember{
		UserID:  memberID,
		GroupID: groupID,
	}

	err := g.groupsClient.RemoveGroupMember(r.Context(), groupMember)
	if err != nil {
		if errors.Is(err, groupsclient.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.log.Error("error while removing group member", slog.Any("groupMember", groupMember), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Group) ChangeMemberRole(w http.ResponseWriter, r *http.Request) {
	groupID := GetGroupID(r)
	memberID := GetGroupMemberID(r)
	newRole := GetNewRole(r)

	groupMember := &models.GroupMember{
		GroupID: groupID,
		UserID:  memberID,
		Role:    newRole,
	}

	err := g.groupsClient.ChangeMemberRole(r.Context(), groupMember)
	if err != nil {
		if errors.Is(err, groupsclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, groupsclient.ErrPermissionDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, groupsclient.ErrBadRoleParam) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		g.log.Error("error while changing member role", slog.Any("groupMember", groupMember), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
