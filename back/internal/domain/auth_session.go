package domain

import "time"

type AuthSession struct {
	ID        int64
	UserID    int64
	TokenHash [32]byte
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt time.Time
}
