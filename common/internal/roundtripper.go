package internal

import (
	"context"
	"net/http"
	"time"

	"github.com/anz-bank/sysl-go/log"
)

type loggingRoundtripper struct {
	ctx  context.Context
	base http.RoundTripper
}

func NewLoggingRoundTripper(ctx context.Context, base http.RoundTripper) http.RoundTripper {
	return &loggingRoundtripper{
		ctx:  ctx,
		base: base,
	}
}

func (t *loggingRoundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	var resp *http.Response
	reqLogger, _ := NewRequestLogger(t.ctx, req)
	defer func() {
		reqLogger.LogResponse(resp)
	}()

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if val, ok := req.Context().Value(unclosedResponseBodyMonitorContextKey{}).(*unclosedResponseBodyMonitor); ok {
		val.addResponse(resp)
	}

	reqTime := time.Since(start)

	ctx := initCommonLogFields(t.ctx, resp.StatusCode, reqTime, resp.Request)

	log.Info(ctx, "Backend request completed")
	return resp, nil
}
