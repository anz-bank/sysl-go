package common

import (
	"context"
	"net/http"

	"github.com/anz-bank/sysl-go/common/internal"

	"github.com/anz-bank/pkg/log"
	"github.com/sirupsen/logrus"
)

// Deprecated: Use ServerParams.WithPkgLogger instead.
func GetLogEntryFromContext(ctx context.Context) *logrus.Entry {
	core := ctx.Value(coreRequestContextKey{})
	if core == nil {
		return nil
	}
	return core.(coreRequestContext).entry
}

// Deprecated: Use ServerParams.WithPkgLogger instead.
func GetLoggerFromContext(ctx context.Context) *logrus.Logger {
	core := ctx.Value(coreRequestContextKey{})
	if core == nil {
		return nil
	}
	return core.(coreRequestContext).logger
}

func NewLoggingRoundTripper(name string, base http.RoundTripper) http.RoundTripper {
	// temporary pass-through to get the real round tripper from the request context
	return &tempRoundtripper{name, base}
}

type coreRequestContext struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

type reqHeaderContext struct {
	header http.Header
}

type respHeaderAndStatusContext struct {
	header http.Header
	status int
}

type RestResult struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}

// LoggerToContext creates a new context containing the logger.
// Deprecated: Use ServerParams.WithPkgLogger instead.
func LoggerToContext(ctx context.Context, logger *logrus.Logger, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, coreRequestContextKey{}, coreRequestContext{logger, entry})
}

// RequestHeaderToContext creates a new context containing the request header.
func RequestHeaderToContext(ctx context.Context, header http.Header) context.Context {
	return context.WithValue(ctx, reqHeaderContextKey{}, &reqHeaderContext{header})
}

// RequestHeaderFromContext retrieves the request header from the context.
func RequestHeaderFromContext(ctx context.Context) http.Header {
	reqHeader := getReqHeaderContext(ctx)

	if reqHeader == nil {
		return nil
	}
	return reqHeader.header
}

// RespHeaderAndStatusToContext creates a new context containing the response header and status.
func RespHeaderAndStatusToContext(ctx context.Context, header http.Header, status int) context.Context {
	return context.WithValue(ctx, respHeaderAndStatusContextKey{}, &respHeaderAndStatusContext{header, status})
}

// RespHeaderAndStatusFromContext retrieves response header and status from the context.
func RespHeaderAndStatusFromContext(ctx context.Context) (header http.Header, status int) {
	respHeaderAndStatus := getRespHeaderAndStatusContext(ctx)

	if respHeaderAndStatus == nil {
		return nil, http.StatusOK
	}

	header = respHeaderAndStatus.header
	status = respHeaderAndStatus.status
	return
}

func UpdateResponseStatus(ctx context.Context, status int) error {
	respHeaderAndStatus := getRespHeaderAndStatusContext(ctx)

	if respHeaderAndStatus == nil {
		return CreateError(ctx, InternalError, "response status not in context", nil)
	}
	respHeaderAndStatus.status = status
	return nil
}

func CoreRequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = log.With(traceIDLogField, GetTraceIDFromContext(ctx)).Onto(ctx)

		ctx = internal.AddResponseBodyMonitorToContext(ctx)
		defer internal.CheckForUnclosedResponses(ctx)
		reqLogger, entry := internal.NewRequestLogger(ctx, r)
		w = reqLogger.ResponseWriter(w)
		defer reqLogger.FlushLog()

		r = r.WithContext(ctx)

		tl := internal.NewRequestTimer(w, r)
		w = tl.RespWrapper
		defer tl.Log(entry)

		next.ServeHTTP(w, r)
	})
}

type coreRequestContextKey struct{}

type reqHeaderContextKey struct{}
type respHeaderAndStatusContextKey struct{}

func getReqHeaderContext(ctx context.Context) *reqHeaderContext {
	reqHeaderCtx := ctx.Value(reqHeaderContextKey{})
	if reqHeaderCtx == nil {
		return nil
	}
	return reqHeaderCtx.(*reqHeaderContext)
}

func getRespHeaderAndStatusContext(ctx context.Context) *respHeaderAndStatusContext {
	respHeaderAndStatusCtx := ctx.Value(respHeaderAndStatusContextKey{})
	if respHeaderAndStatusCtx == nil {
		return nil
	}
	return respHeaderAndStatusCtx.(*respHeaderAndStatusContext)
}

type tempRoundtripper struct {
	name string
	base http.RoundTripper
}

func (t *tempRoundtripper) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := log.With("Downstream", t.name).Onto(r.Context())
	return internal.NewLoggingRoundTripper(ctx, t.base).RoundTrip(r)
}

type RestResultContextKey struct{}

// ProvisionRestResult provisions within the context the ability to retrieve the
// result of a rest request.
func ProvisionRestResult(ctx context.Context) context.Context {
	return context.WithValue(ctx, RestResultContextKey{}, &RestResult{})
}

// GetRestResult gets the result of the most recent rest request. The context
// must be provisioned prior to the request taking place with a call to
// ProvisionRestResult.
func GetRestResult(ctx context.Context) *RestResult {
	raw := ctx.Value(RestResultContextKey{})
	if raw == nil {
		return nil
	}
	return raw.(*RestResult)
}
