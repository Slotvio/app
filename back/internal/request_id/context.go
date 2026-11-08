package requestid

import (
	"context"
	"errors"
)

var (
	errRequestIDNotFound = errors.New("request_id not found in context")
)

type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestIDFromContext(ctx context.Context) (string, error) {
	id, ok := ctx.Value(requestIDKey{}).(string)
	if !ok {
		return "", errRequestIDNotFound
	}
	return id, nil
}
