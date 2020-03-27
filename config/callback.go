package config

import (
	"context"
	"net/http"
	"time"

	"github.com/anz-bank/sysl-go/common"
	"github.com/go-chi/chi"
)

// GenCallback callbacks used by the generated code
type GenCallback interface {
	AddMiddleware(ctx context.Context, r chi.Router)
	BasePath() string
	Config() UpstreamConfig
	HandleError(ctx context.Context, w http.ResponseWriter, kind common.Kind, message string, cause error)
	DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc)
}

type Callback struct {
	UpstreamTimeout   time.Duration
	DownstreamTimeout time.Duration
	ErrorHandler      ErrorHandler
	RouterBasePath    string
	UpstreamConfig    UpstreamConfig
}

// ErrorHandler struct
type ErrorHandler struct{}

// DownStreamTimeoutContext func
func (c Callback) DownstreamTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, c.DownstreamTimeout)
}

// Config func
func (c Callback) Config() UpstreamConfig {
	return c.UpstreamConfig
}

// HandleError func
func (c Callback) HandleError(ctx context.Context, w http.ResponseWriter, kind common.Kind, message string, cause error) {
	c.ErrorHandler.HandleError(ctx, w, common.CreateError(ctx, kind, message, cause))
}

// AddMiddleware func
func (c Callback) AddMiddleware(ctx context.Context, r chi.Router) {
	r.Use(
		common.Timeout(ctx, c.UpstreamTimeout, http.HandlerFunc(c.timeoutHandler)),
	)
}

// BasePath func
func (c Callback) BasePath() string {
	if c.RouterBasePath == "" {
		return "/"
	}
	return c.RouterBasePath
}

// Timeout handler
func (c Callback) timeoutHandler(w http.ResponseWriter, r *http.Request) {
	err := common.CreateError(r.Context(), common.InternalError, "timeout expired while processing response", nil)
	c.ErrorHandler.HandleError(r.Context(), w, err)
}
