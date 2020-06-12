package common

import (
	"context"
	"net/http"

	"github.com/anz-bank/sysl-go/common/internal"

	"github.com/anz-bank/pkg/log"
	"github.com/google/uuid"
)

type traceabilityContextKey struct{}

type requestID struct {
	id          uuid.UUID
	wasProvided bool
}

const traceIDLogField = "traceid"

// Injects a traceId UUID into the request context
func TraceabilityMiddleware(ctx context.Context) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			val, err := uuid.Parse(r.Header.Get("RequestID"))
			if err != nil {
				log.Info(internal.InitFieldsFromRequest(r).Onto(ctx), "Incoming request with invalid or missing RequestID header, filled traceid with new UUID instead")
				r = r.WithContext(AddTraceIDToContext(r.Context(), uuid.New(), false))
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
