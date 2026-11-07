package requestid

import (
	"context"
	"errors"
)

var (
	errInvalidIDType = errors.New("request_id isn't a string")
)

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, "request_id", id)
}

func RequestIDFromContext(ctx context.Context) (string, error) {
	id, ok := ctx.Value("request_id").(string)

	if !ok {
		return "", errInvalidIDType
	}

	return id, nil
}
