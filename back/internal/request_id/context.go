package requestid

import (
	"context"
	"errors"
)

var (
	errInvalidIDType = errors.New("request_id isn't a string")
)

type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func RequestIDFromContext(ctx context.Context) (string, error) {
	id, ok := ctx.Value(requestIDKey{}).(string)
	if !ok {
		return "", errors.New("request_id not found in context")
	}
	return id, nil
}
