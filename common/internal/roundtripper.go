package internal

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type loggingRoundtripper struct {
	logentry *logrus.Entry
	base     http.RoundTripper
}

func NewLoggingRoundTripper(logentry *logrus.Entry, base http.RoundTripper) http.RoundTripper {
	return &loggingRoundtripper{
		logentry: logentry,
		base:     base,
	}
}

func (t *loggingRoundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	var resp *http.Response
	reqLogger, _ := NewRequestLogger(t.logentry, req)
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

	fields := initCommonLogFields(resp.StatusCode, reqTime, resp.Request)

	entry := t.logentry.WithFields(fields).WithFields(logrus.Fields{
		"logger": "common/internal/roundtripper.go",
		"func":   "RoundTrip()",
	})
	entry.Info("Backend request completed")

	return resp, nil
}
