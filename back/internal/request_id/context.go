package requestid

import (
	"context"
	"errors"
)

type contextKey string

const requestIDKey contextKey = "request_id"

var (
	errInvalidIDType = errors.New("request_id isn't a string")
)

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) (string, error) {
	id, ok := ctx.Value(requestIDKey).(string)

	if !ok {
		return "", errInvalidIDType
	}

	return id, nil
}
