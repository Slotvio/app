package auth

import (
	"context"
)

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value("user_id").(int64)
	return userID, ok
}

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, "user_id", userID)
}
