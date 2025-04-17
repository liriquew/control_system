package groups

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/liriquew/control_system/internal/api_handlers/auth"
	groupsclient "github.com/liriquew/control_system/internal/grpc/clients/groups"
)

type GroupID struct{}
type MemberID struct{}

type GroupsMiddleware struct {
	log    *slog.Logger
	client groupsclient.GroupsClient
}

func NewAuthMiddleware(log *slog.Logger, client groupsclient.GroupsClient) *GroupsMiddleware {
	return &GroupsMiddleware{
		log:    log,
		client: client,
	}
}

func GetGroupID(r *http.Request) int64 {
	id, _ := r.Context().Value(GroupID{}).(int64)
	return id
}

func (g *GroupsMiddleware) CheckAdminPermission(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		groupID, err := strconv.ParseInt(chi.URLParam(r, "groupID"), 10, 64)
		g.log.Debug("groupID in middleware", slog.Int64("groupID", groupID))
		if err != nil {
			http.Error(w, "groupID path param required", http.StatusBadRequest)
			return
		}
		userID := auth.GetUserID(r)
		g.log.Debug("userID in middleware", slog.Int64("userID", userID))

		err = g.client.CheckAdminPermission(r.Context(), userID, groupID)
		if err != nil {
			if errors.Is(err, groupsclient.ErrPermissionDenied) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		ctx := context.WithValue(r.Context(), GroupID{}, groupID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (g *GroupsMiddleware) CheckEditorPermission(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		groupID, err := strconv.ParseInt(chi.URLParam(r, "groupID"), 10, 64)
		g.log.Debug("groupID in middleware", slog.Int64("groupID", groupID))
		if err != nil {
			http.Error(w, "groupID path param required", http.StatusBadRequest)
			return
		}
		userID := auth.GetUserID(r)
		g.log.Debug("userID in middleware", slog.Int64("userID", userID))

		err = g.client.CheckEditorPermission(r.Context(), userID, groupID)
		if err != nil {
			if errors.Is(err, groupsclient.ErrPermissionDenied) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		ctx := context.WithValue(r.Context(), GroupID{}, groupID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (g *GroupsMiddleware) CheckMemberPermission(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		groupID, err := strconv.ParseInt(chi.URLParam(r, "groupID"), 10, 64)
		g.log.Debug("groupID in middleware", slog.Int64("groupID", groupID))
		if err != nil {
			http.Error(w, "groupID path param required", http.StatusBadRequest)
			return
		}
		userID := auth.GetUserID(r)
		g.log.Debug("userID in middleware", slog.Int64("userID", userID))

		err = g.client.CheckMemberPermission(r.Context(), userID, groupID)
		if err != nil {
			if errors.Is(err, groupsclient.ErrPermissionDenied) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		ctx := context.WithValue(r.Context(), GroupID{}, groupID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (g *GroupsMiddleware) ExtractMemberID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		groupID, err := strconv.ParseInt(chi.URLParam(r, "memberID"), 10, 64)
		if err != nil {
			http.Error(w, "memberID path param required", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), MemberID{}, groupID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetGroupMemberID(r *http.Request) int64 {
	id, _ := r.Context().Value(MemberID{}).(int64)
	return id
}

func GetNewRole(r *http.Request) string {
	newRole := r.URL.Query().Get("newRole")
	return newRole
}

func GetPadding(r *http.Request) int64 {
	padding, _ := strconv.ParseInt(r.URL.Query().Get("padding"), 10, 64)
	return padding
}
