package jwtgrpc

import (
	"context"
)

type JWTKey struct{}

func FromContext(ctx context.Context) string {
	val, ok := ctx.Value(JWTKey{}).(string)
	if !ok {
		return ""
	}
	return val
}
