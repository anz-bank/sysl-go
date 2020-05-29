package internal

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
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

func (r RequestTimer) Log(entry *logrus.Entry) {
	latency := time.Since(r.start)
	status := r.RespWrapper.Status()

	fields := initCommonLogFields(status, latency, r.r)
	switch {
	case status < 400:
		entry.WithFields(fields).Info("Request completed")
	default:
		entry.WithFields(fields).Error("Request completed with error status")
	}
}

func initCommonLogFields(status int, reqTime time.Duration, req *http.Request) logrus.Fields {
	fields := InitFieldsFromRequest(req)
	fields["status"] = status
	fields["took"] = reqTime
	fields["latency"] = reqTime.Nanoseconds()
	return fields
}

const (
	distributedTraceIDName      = "X-B3-Traceid"
	distributedSpanIDName       = "X-B3-Spanid"
	distributedParentSpanIDName = "X-B3-ParentSpanId"
)

var distributedTracingIDs = []string{distributedTraceIDName, distributedSpanIDName, distributedParentSpanIDName}

func InitFieldsFromRequest(req *http.Request) logrus.Fields {
	fields := logrus.Fields{
		"remote":  req.RemoteAddr,
		"request": req.URL,
		"method":  req.Method,
	}
	addDistributedTracingFields(req.Header, fields)
	return fields
}

func addDistributedTracingFields(header http.Header, fields logrus.Fields) {
	for _, name := range distributedTracingIDs {
		if val := header.Get(name); val != "" {
			fields[name] = val
		}
	}
}
