package auth

import "context"

func UserIDFromContext(ctx context.Context) (int64, bool) {
	return 1, true
}
