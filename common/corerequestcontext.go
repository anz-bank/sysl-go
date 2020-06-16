package common

import (
	"context"
	"net/http"

	"github.com/anz-bank/sysl-go/common/internal"

	"github.com/anz-bank/pkg/log"
	"github.com/sirupsen/logrus"
)

// Deprecated: Use ServerParams.WithPkgLogger instead
func GetLogEntryFromContext(ctx context.Context) *logrus.Entry {
	return getCoreContext(ctx).entry
}

// Deprecated: Use ServerParams.WithPkgLogger instead
func GetLoggerFromContext(ctx context.Context) *logrus.Logger {
	return getCoreContext(ctx).logger
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

// LoggerToContext create a new context containing the logger
// Deprecated: Use ServerParams.WithPkgLogger instead
func LoggerToContext(ctx context.Context, logger *logrus.Logger, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, coreRequestContextKey{}, &coreRequestContext{logger, entry})
}

// RequestHeaderToContext create a new context containing the request header
func RequestHeaderToContext(ctx context.Context, header http.Header) context.Context {
	return context.WithValue(ctx, reqHeaderContextKey{}, &reqHeaderContext{header})
}

// RequestHeaderFromContext retrieve the request header from the context
func RequestHeaderFromContext(ctx context.Context) http.Header {
	reqHeader := getReqHeaderContext(ctx)

	if reqHeader == nil {
		return nil
	}
	return reqHeader.header
}

// RespHeaderAndStatusToContext create a new context containing the response header and status
func RespHeaderAndStatusToContext(ctx context.Context, header http.Header, status int) context.Context {
	return context.WithValue(ctx, respHeaderAndStatusContextKey{}, &respHeaderAndStatusContext{header, status})
}

// RespHeaderAndStatusFromContext retrieve response header and status from the context
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

func CoreRequestContextMiddleware(logger ...*logrus.Logger) func(next http.Handler) http.Handler {
	var ctxlogger *logrus.Logger
	if len(logger) == 1 {
		ctxlogger = logger[0]
	}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if ctxlogger != nil {
				ctx = LoggerToContext(ctx, ctxlogger, nil)
			}
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
		}
		return http.HandlerFunc(fn)
	}
}

type coreRequestContextKey struct{}

func getCoreContext(ctx context.Context) *coreRequestContext {
	return ctx.Value(coreRequestContextKey{}).(*coreRequestContext)
}

type reqHeaderContextKey struct{}
type respHeaderAndStatusContextKey struct{}

func getReqHeaderContext(ctx context.Context) *reqHeaderContext {
	if ctx.Value(reqHeaderContextKey{}) == nil {
		return nil
	}
	return ctx.Value(reqHeaderContextKey{}).(*reqHeaderContext)
}

func getRespHeaderAndStatusContext(ctx context.Context) *respHeaderAndStatusContext {
	if ctx.Value(respHeaderAndStatusContextKey{}) == nil {
		return nil
	}
	return ctx.Value(respHeaderAndStatusContextKey{}).(*respHeaderAndStatusContext)
}

type tempRoundtripper struct {
	name string
	base http.RoundTripper
}

func (t *tempRoundtripper) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := log.With("Downsteam", t.name).Onto(r.Context())
	return internal.NewLoggingRoundTripper(ctx, t.base).RoundTrip(r)
}
