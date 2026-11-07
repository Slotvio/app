package middleware

import "context"

type Authenticator interface {
	Authenticate(ctx context.Context, initData string) (Principal, error)
}

type Principal struct {
	UserID     int64
	TelegramID int64
}
