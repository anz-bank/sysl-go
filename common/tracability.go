package common

import (
	"context"
	"net/http"

	"github.com/anz-bank/sysl-go/common/internal"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type traceabilityContextKey struct{}

type requestID struct {
	id string // This appears to be a UUID in both the android and IOS clients, however it is 'string' in the swagger :/
}

const traceIDLogField = "traceid"

// Injects a traceId UUID into the request context
func TraceabilityMiddleware(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			reqid := r.Header.Get("RequestID")
			if reqid == "" {
				reqid = uuid.New().String()
				logger.WithFields(internal.InitFieldsFromRequest(r)).
					WithField(traceIDLogField, reqid).
					Warnf("Incoming request without RequestID header, filled traceid with new UUID instead")
			}

			r = r.WithContext(context.WithValue(r.Context(), traceabilityContextKey{}, &requestID{reqid}))

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func GetTraceIDFromContext(ctx context.Context) string {
	val, ok := ctx.Value(traceabilityContextKey{}).(*requestID)
	if ok {
		return val.id
	}
	return ""
}
