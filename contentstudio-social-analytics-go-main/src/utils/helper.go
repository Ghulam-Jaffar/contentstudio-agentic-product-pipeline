package utils

import (
	"context"
)

type CtxKey string

const requestIDKey CtxKey = "request_id"

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func GetRequestID(ctx context.Context) string {
	return ctx.Value(requestIDKey).(string)
}
