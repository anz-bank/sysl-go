package common

import (
	"context"
	"net/http"
	"time"

	"github.com/anz-bank/sysl-go/config"

	"github.com/anz-bank/sysl-go/validator"
	"github.com/go-chi/chi"
)

func DefaultCallback() Callback {
	return Callback{
		DownstreamTimeout: 120 * time.Second,
		RouterBasePath:    "/",
		UpstreamConfig:    Config{},
	}
}

type Callback struct {
	DownstreamTimeout time.Duration
	RouterBasePath    string
	UpstreamConfig    validator.Validator
	MapErrorFunc      func(ctx context.Context, err error) *HTTPError                        // MapErrorFunc may be left nil to use default behaviour.
	WriteErrorFunc    func(ctx context.Context, w http.ResponseWriter, httpError *HTTPError) // WriteError may be left nil to use default behaviour.
	AddMiddlewareFunc func(ctx context.Context, r chi.Router)                                // AddMiddlewareFunc may be left nil to use default behaviour.
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

func (g Callback) WriteError(ctx context.Context, w http.ResponseWriter, httpError *HTTPError) {
	if g.WriteErrorFunc == nil {
		httpError.WriteError(ctx, w)
	} else {
		g.WriteErrorFunc(ctx, w, httpError)
	}
}

func NewCallbackV3(
	config *config.GenCodeConfig,
	downstreamTimeOut time.Duration,
	mapError func(ctx context.Context, err error) *HTTPError,
	writeError func(ctx context.Context, w http.ResponseWriter, httpError *HTTPError),
	addMiddleware func(ctx context.Context, r chi.Router),
) Callback {
	// construct the rest configuration (aka. gen callback)
	return Callback{
		DownstreamTimeout: downstreamTimeOut,
		RouterBasePath:    config.Upstream.HTTP.BasePath,
		UpstreamConfig:    &config.Upstream,
		MapErrorFunc:      mapError,
		WriteErrorFunc:    writeError,
		AddMiddlewareFunc: addMiddleware,
	}
}

// NewCallbackV2 is deprecated, prefer NewCallbackV3.
func NewCallbackV2(
	config *config.GenCodeConfig,
	downstreamTimeOut time.Duration,
	mapError func(ctx context.Context, err error) *HTTPError,
	addMiddleware func(ctx context.Context, r chi.Router),
) Callback {
	// construct the rest configuration (aka. gen callback)
	return Callback{
		DownstreamTimeout: downstreamTimeOut,
		RouterBasePath:    config.Upstream.HTTP.BasePath,
		UpstreamConfig:    &config.Upstream,
		MapErrorFunc:      mapError,
		AddMiddlewareFunc: addMiddleware,
	}
}

// NewCallback is deprecated, prefer NewCallbackV2.
func NewCallback(
	config *config.GenCodeConfig,
	downstreamTimeOut time.Duration,
	mapError func(ctx context.Context, err error) *HTTPError,
) Callback {
	return NewCallbackV2(
		config,
		downstreamTimeOut,
		mapError,
		nil,
	)
}
