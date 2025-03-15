package groups

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type GroupID struct{}
type MemberID struct{}

func ExtractGroupID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		groupID, err := strconv.ParseInt(chi.URLParam(r, "groupID"), 10, 64)
		if err != nil {
			http.Error(w, "groupID path param required", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), GroupID{}, groupID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ExtractMemberID(next http.Handler) http.Handler {
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
