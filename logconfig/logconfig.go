package logconfig

import (
	"context"
)

type isVerboseLoggingKey struct{}
type isVerboseLogging struct {
	Flag bool
}

// Get whether or not sysl-go internals log additional debug/verbose level information.
func IsVerboseLogging(ctx context.Context) bool {
	v, ok := ctx.Value(isVerboseLoggingKey{}).(*isVerboseLogging)
	if ok {
		return v.Flag
	}
	return false
}

// Set against the context whether or not sysl-go internals log additional debug/verbose level information.
func SetVerboseLogging(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, isVerboseLoggingKey{},
		&isVerboseLogging{
			Flag: verbose,
		})
}
