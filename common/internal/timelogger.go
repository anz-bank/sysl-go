package internal

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/anz-bank/sysl-go/log"

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

	ctx = initCommonLogFields(ctx, status, latency, r.r)
	switch {
	case status < 400:
		log.Info(ctx, "Request completed")
	default:
		log.Error(ctx, fmt.Errorf("invalid status: %d", status), "request completed with error status")
	}
}

func initCommonLogFields(ctx context.Context, status int, reqTime time.Duration, req *http.Request) context.Context {
	ctx = InitFieldsFromRequest(ctx, req)
	ctx = log.WithInt(ctx, "status", status)
	ctx = log.WithDuration(ctx, "took", reqTime)
	ctx = log.WithStr(ctx, "latency", strconv.FormatInt(reqTime.Nanoseconds(), 10))
	return ctx
}

const (
	distributedTraceIDName      = "X-B3-Traceid"
	distributedSpanIDName       = "X-B3-Spanid"
	distributedParentSpanIDName = "X-B3-ParentSpanId"
)

var distributedTracingIDs = []string{distributedTraceIDName, distributedSpanIDName, distributedParentSpanIDName}

func InitFieldsFromRequest(ctx context.Context, req *http.Request) context.Context {
	ctx = distributedTracingFields(ctx, req.Header)
	ctx = log.WithStr(ctx, "remote", req.RemoteAddr)
	ctx = log.WithStr(ctx, "request", req.URL.String())
	ctx = log.WithStr(ctx, "method", req.Method)
	return ctx
}

func distributedTracingFields(ctx context.Context, header http.Header) context.Context {
	for _, name := range distributedTracingIDs {
		if val := header.Get(name); val != "" {
			ctx = log.WithStr(ctx, name, val)
		}
	}
	return ctx
}
