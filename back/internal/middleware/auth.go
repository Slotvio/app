package middleware

import (
	"context"
	"net/http"
)

type Authenticator interface {
	Authenticate(ctx context.Context, initData string) (Principal, error)
}

type Principal struct {
	UserID     int64
	TelegramID int64
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
