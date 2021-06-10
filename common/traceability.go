package common

import (
	"context"
	"net/http"

	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/log"

	"github.com/anz-bank/sysl-go/common/internal"

	"github.com/google/uuid"
)

type traceabilityContextKey struct{}

type requestID struct {
	id          uuid.UUID
	wasProvided bool
}

const traceIDLogField = "traceid"
const defaultIncomingHeaderForID = "RequestID"

// Injects a traceId UUID into the request context.
func TraceabilityMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		val, err := uuid.Parse(r.Header.Get(getIncomingHeaderForID(ctx)))
		if err != nil {
			log.Info(internal.InitFieldsFromRequest(ctx, r), "Incoming request with invalid or missing RequestID header, filled traceid with new UUID instead")
			r = r.WithContext(AddTraceIDToContext(r.Context(), uuid.New(), false))
		} else {
			r = r.WithContext(AddTraceIDToContext(r.Context(), val, true))
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
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

func getIncomingHeaderForID(ctx context.Context) string {
	ret := defaultIncomingHeaderForID
	cfg := config.GetDefaultConfig(ctx)
	if cfg != nil && cfg.Library.Trace.IncomingHeaderForID != "" {
		ret = cfg.Library.Trace.IncomingHeaderForID
	}

	return ret
}
