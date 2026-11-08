package requestid

import (
	"context"
	"errors"
)

type RequestIDType string

var (
	errInvalidIDType               = errors.New("request_id isn't a string")
	requestIDKey     RequestIDType = "request_id"
)

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) (RequestIDType, error) {
	id, ok := ctx.Value(requestIDKey).(RequestIDType)

	if !ok {
		return "", errInvalidIDType
	}

	return id, nil
}
