package common

import (
	"context"
	"net/http"
	"time"

	"github.com/anz-bank/sysl-go/validator"
	"github.com/go-chi/chi"
)

func DefaultCallback() Callback {
	return Callback{
		UpstreamTimeout:   120 * time.Second,
		DownstreamTimeout: 120 * time.Second,
		RouterBasePath:    "/",
		UpstreamConfig:    Config{},
	}
}

type Callback struct {
	UpstreamTimeout   time.Duration
	DownstreamTimeout time.Duration
	RouterBasePath    string
	UpstreamConfig    validator.Validator
	MapErrorFunc      func(ctx context.Context, err error) *HTTPError // MapErrorFunc may be left nil to use default behaviour.
	AddMiddlewareFunc func(ctx context.Context, r chi.Router)         // AddMiddlewareFunc may be left nil to use default behaviour.
}

type Config struct{}

func (c Config) Validate() error {
	return nil
}

func (g Callback) AddMiddleware(ctx context.Context, r chi.Router) {
	if g.AddMiddlewareFunc != nil {
		g.AddMiddlewareFunc(ctx, r)
	}
}

func (g Callback) BasePath() string {
	return g.RouterBasePath
}

func (g Callback) Config() interface{} {
	return g.UpstreamConfig
}

func (g Callback) HandleError(ctx context.Context, w http.ResponseWriter, kind Kind, message string, cause error) {
	se := CreateError(ctx, kind, message, cause)
	g.MapError(ctx, se).WriteError(ctx, w)
}

func (g Callback) DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, g.DownstreamTimeout)
}

// MapError maps an error to an HTTPError in instances where custom error mapping is required.
// Return nil to perform default error mapping; defined as:
// 1. CustomError.HTTPError if the original error is a CustomError, otherwise
// 2. common.MapError.
func (g Callback) MapError(ctx context.Context, err error) *HTTPError {
	if g.MapErrorFunc == nil {
		return nil
	}
	return g.MapErrorFunc(ctx, err)
}
