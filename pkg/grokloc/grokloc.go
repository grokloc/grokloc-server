// Package grokloc provides system-wide functionality and values
package grokloc

import (
	"context"

	"github.com/google/uuid"
)

// RequestIDKeyType is used to set the request id in a context
type RequestIDKeyType string

// RequestIDKey is the context key for the request id
const RequestIDKey = "RequestID"

func CtxRequestID(ctx context.Context) string {
	k := RequestIDKeyType(RequestIDKey)
	v := ctx.Value(k)
	if v != nil {
		s, ok := v.(string)
		if !ok {
			panic("cannot assert request id to string")
		}
		return s
	}
	return ""
}

func WithRequestID(ctx context.Context) context.Context {
	return context.WithValue(ctx, RequestIDKeyType(RequestIDKey), uuid.NewString())
}
