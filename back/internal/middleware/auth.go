package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/aleksandrboro/slotvio/internal/auth"
	"github.com/gorilla/mux"
)

type SessionAuthenticator interface {
	AuthenticateSession(ctx context.Context, token string) (userID int64, err error)
}

func AuthMiddleware(a SessionAuthenticator) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := a.AuthenticateSession(r.Context(), parts[1])
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := auth.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
