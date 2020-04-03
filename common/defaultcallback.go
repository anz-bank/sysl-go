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
}

type Config struct{}

func (c Config) Validate() error {
	return nil
}

func (g Callback) AddMiddleware(ctx context.Context, r chi.Router) {
}

func (g Callback) BasePath() string {
	return g.RouterBasePath
}

func (g Callback) Config() validator.Validator {
	return g.UpstreamConfig
}

func (g Callback) HandleError(ctx context.Context, w http.ResponseWriter, kind Kind, message string, cause error) {
	se := CreateError(ctx, kind, message, cause)

	httpError := MapError(ctx, se)

	httpError.WriteError(ctx, w)
}

func (g Callback) DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, g.DownstreamTimeout)
}
