package internal

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/anz-bank/pkg/log"
	"github.com/go-chi/chi/middleware"
)

type RequestTimer struct {
	start       time.Time
	r           *http.Request
	RespWrapper middleware.WrapResponseWriter
}

func NewRequestTimer(w http.ResponseWriter, r *http.Request) RequestTimer {
	rt := RequestTimer{
		start:       time.Now(),
		r:           r,
		RespWrapper: middleware.NewWrapResponseWriter(w, r.ProtoMajor),
	}
	return rt
}

func (r RequestTimer) Log(ctx context.Context) {
	latency := time.Since(r.start)
	status := r.RespWrapper.Status()

	fields := initCommonLogFields(status, latency, r.r)
	switch {
	case status < 400:
		log.Info(fields.Onto(ctx), "Request completed")
	default:
		log.Error(fields.Onto(ctx), errors.New("request completed with error status"))
	}
}

func initCommonLogFields(status int, reqTime time.Duration, req *http.Request) log.Fields {
	return InitFieldsFromRequest(req).
		With("status", status).
		With("took", reqTime).
		With("latency", reqTime.Nanoseconds())
}

const (
	distributedTraceIDName      = "X-B3-Traceid"
	distributedSpanIDName       = "X-B3-Spanid"
	distributedParentSpanIDName = "X-B3-ParentSpanId"
)

var distributedTracingIDs = []string{distributedTraceIDName, distributedSpanIDName, distributedParentSpanIDName}

func InitFieldsFromRequest(req *http.Request) log.Fields {
	return distributedTracingFields(req.Header).
		With("remote", req.RemoteAddr).
		With("request", req.URL).
		With("method", req.Method)
}

func distributedTracingFields(header http.Header) log.Fields {
	fields := log.Fields{}
	for _, name := range distributedTracingIDs {
		if val := header.Get(name); val != "" {
			fields = fields.Chain(log.With(name, val))
		}
	}
	return fields
}
