package internal

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/anz-bank/pkg/log"
	"github.com/anz-bank/sysl-go/logconfig"

	"github.com/go-chi/chi/middleware"
)

type RequestLogger interface {
	LogResponse(response *http.Response) // Calls Flush() automatically

	ResponseWriter(http.ResponseWriter) http.ResponseWriter
	FlushLog() // Must be called if using the ResponseWriter() func
}

type httpData struct {
	body   bytes.Buffer
	header http.Header
}

type requestLogger struct {
	ctx        context.Context
	req        httpData
	resp       httpData
	protoMajor int
	rw         http.ResponseWriter
	flushed    bool
}

func (r *requestLogger) LogResponse(resp *http.Response) {
	if resp != nil {
		b, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(b))
		r.resp.body.Write(b)
		r.resp.header = resp.Header
	}

	r.FlushLog()
}

func (r *requestLogger) ResponseWriter(base http.ResponseWriter) http.ResponseWriter {
	rw := middleware.NewWrapResponseWriter(base, r.protoMajor)
	rw.Tee(&r.resp.body)

	r.rw = rw

	return rw
}

func (r *requestLogger) FlushLog() {
	if r.flushed {
		log.Info(r.ctx, "Already flushed the request")
		return
	}
	r.flushed = true
	if r.rw != nil {
		r.resp.header = r.rw.Header()
	}

	fields := log.With("logger", "common/internal/requestlogger.go").With("func", "FlushLog()")

	reqBody := r.req.body.String()
	fields.Debugf(r.ctx, "Request: header - %s\nbody[len:%v]: - %s", r.req.header, len(reqBody), reqBody)
	respBody := r.resp.body.String()
	fields.Debugf(r.ctx, "Response: header - %s\nbody[len:%v]: - %s", r.resp.header, len(respBody), respBody)
}

type nopLogger struct{}

func (r *nopLogger) LogResponse(_ *http.Response)                                {}
func (r *nopLogger) ResponseWriter(base http.ResponseWriter) http.ResponseWriter { return base }
func (r *nopLogger) FlushLog()                                                   {}

func NewRequestLogger(ctx context.Context, req *http.Request) (RequestLogger, context.Context) {
	if logconfig.IsVerboseLogging(ctx) {
		l := &requestLogger{
			ctx:        InitFieldsFromRequest(req).Onto(ctx),
			protoMajor: req.ProtoMajor,
		}
		l.req.header = req.Header.Clone()
		if req.Body != nil && req.Method != http.MethodGet {
			b, _ := ioutil.ReadAll(req.Body)
			_ = req.Body.Close()
			req.Body = ioutil.NopCloser(bytes.NewBuffer(b))
			l.req.body.Write(b)
		}
		return l, l.ctx
	}
	return &nopLogger{}, ctx
}
