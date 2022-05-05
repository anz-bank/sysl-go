package common

import (
	"context"
	"net/http"

	"github.com/anz-bank/sysl-go/common/internal"
	"github.com/anz-bank/sysl-go/log"

	"github.com/sirupsen/logrus"
)

// Deprecated: Use log.GetLogger.
func GetLogEntryFromContext(ctx context.Context) *logrus.Entry {
	return log.GetLogrusLogEntryFromContext(ctx)
}

// Deprecated: Use log.GetLogger.
func GetLoggerFromContext(ctx context.Context) *logrus.Logger {
	return log.GetLogrusLoggerFromContext(ctx)
}

func NewLoggingRoundTripper(name string, base http.RoundTripper) http.RoundTripper {
	// temporary pass-through to get the real round tripper from the request context
	return &tempRoundtripper{name, base}
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

// Deprecated: Use log.GetLogger.
func LoggerToContext(ctx context.Context, logger *logrus.Logger, entry *logrus.Entry) context.Context {
	return log.LogrusLoggerToContext(ctx, logger, entry)
}

// RequestHeaderToContext creates a new context containing the request header.
func RequestHeaderToContext(ctx context.Context, header http.Header) context.Context {
	return context.WithValue(ctx, reqHeaderContextKey{}, &reqHeaderContext{header.Clone()})
}

// RequestHeaderFromContext retrieves the request header from the context.
func RequestHeaderFromContext(ctx context.Context) http.Header {
	reqHeader := getReqHeaderContext(ctx)

	if reqHeader == nil {
		return nil
	}
	return reqHeader.header.Clone()
}

// RespHeaderAndStatusToContext creates a new context containing the response header and status.
func RespHeaderAndStatusToContext(ctx context.Context, header http.Header, status int) context.Context {
	return context.WithValue(ctx, respHeaderAndStatusContextKey{}, &respHeaderAndStatusContext{header.Clone(), status})
}

// RespHeaderAndStatusFromContext retrieves response header and status from the context.
func RespHeaderAndStatusFromContext(ctx context.Context) (header http.Header, status int) {
	respHeaderAndStatus := getRespHeaderAndStatusContext(ctx)

	if respHeaderAndStatus == nil {
		return nil, http.StatusOK
	}

	header = respHeaderAndStatus.header.Clone()
	status = respHeaderAndStatus.status
	return
}

// AppendToRespHeader will add custom fields to the response header.
func AppendToRespHeader(ctx context.Context, key string, value string) {
	respHeaderAndStatus := getRespHeaderAndStatusContext(ctx)
	if respHeaderAndStatus != nil {
		respHeaderAndStatus.header.Add(key, value)
	}

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
		ctx = log.WithStr(ctx, traceIDLogField, GetTraceIDFromContext(ctx).String())

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
	ctx := log.WithStr(r.Context(), "Downstream", t.name)
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
