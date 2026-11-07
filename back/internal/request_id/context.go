package requestid

import (
	"context"
	"errors"
)

const (
	typeRequestID = "request_id"
)

var (
	errInvalidIDType = errors.New("request_id isn't a string")
)

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, typeRequestID, id)
}

func RequestIDFromContext(ctx context.Context) (string, error) {
	id, ok := ctx.Value(typeRequestID).(string)

	if !ok {
		return "", errInvalidIDType
	}

	return id, nil
}
