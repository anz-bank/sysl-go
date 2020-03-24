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
	id          uuid.UUID
	wasProvided bool
}

const traceIDLogField = "traceid"

// Injects a traceId UUID into the request context
func TraceabilityMiddleware(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			val, err := uuid.Parse(r.Header.Get("RequestID"))
			if err != nil {
				id := uuid.New()
				logger.WithFields(internal.InitFieldsFromRequest(r)).
					Warn("Incoming request with invalid or missing RequestID header, filled traceid with new UUID instead")
				r = r.WithContext(AddTraceIDToContext(r.Context(), id, false))
			} else {
				r = r.WithContext(AddTraceIDToContext(r.Context(), val, true))
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func GetTraceIDFromContext(ctx context.Context) uuid.UUID {
	val, _ := TryGetTraceIDFromContext(ctx)

	return val
}

func TryGetTraceIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	val, ok := ctx.Value(traceabilityContextKey{}).(*requestID)
	if ok {
		return val.id, val.wasProvided
	}
	return uuid.New(), false
}

func AddTraceIDToContext(ctx context.Context, id uuid.UUID, wasProvided bool) context.Context {
	return context.WithValue(ctx, traceabilityContextKey{}, &requestID{id, wasProvided})
}
