package groups

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/liriquew/control_system/api/internal/api_handlers/auth"
	groupsclient "github.com/liriquew/control_system/api/internal/grpc/clients/groups"
)

type GroupID struct{}
type MemberID struct{}
type PredictedUserID struct{}

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

			w.WriteHeader(http.StatusInternalServerError)
			return
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

			w.WriteHeader(http.StatusInternalServerError)
			return
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

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), GroupID{}, groupID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (g *GroupsMiddleware) CheckPredictedUserMemberPermission(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		predictedUserID, err := strconv.ParseInt(chi.URLParam(r, "predictedUserID"), 10, 64)
		if err != nil {
			http.Error(w, "invalid predictedUserID param", http.StatusBadRequest)
			return
		}

		groupID := GetGroupID(r)
		err = g.client.CheckMemberPermission(r.Context(), predictedUserID, groupID)
		if err != nil {
			if errors.Is(err, groupsclient.ErrPermissionDenied) {
				http.Error(w,
					fmt.Sprintf("user id: %d not in group id: %d", predictedUserID, groupID),
					http.StatusBadRequest,
				)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), PredictedUserID{}, predictedUserID)
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

func GetOffset(r *http.Request) int64 {
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	return offset
}

func GetPredictedUserID(r *http.Request) int64 {
	id, _ := r.Context().Value(PredictedUserID{}).(int64)
	return id
}
