package internal

import (
	"bytes"
	"io/ioutil"
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

	var requestBodyCopy []byte
	if t.logentry.Logger.IsLevelEnabled(logrus.DebugLevel) {
		if req.Body != nil {
			requestBodyCopy, _ = ioutil.ReadAll(req.Body)
			req.Body.Close()
			req.Body = ioutil.NopCloser(bytes.NewBuffer(requestBodyCopy))
		}
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if val, ok := req.Context().Value(unclosedResponseBodyMonitorContextKey{}).(*unclosedResponseBodyMonitor); ok {
		val.addResponse(resp)
	}

	reqTime := time.Since(start)

	fields := initCommonLogFields(resp.StatusCode, reqTime, resp.Request)

	entry := t.logentry.WithFields(fields)
	entry.Info("Backend request completed")

	if t.logentry.Logger.IsLevelEnabled(logrus.DebugLevel) {
		rspbody, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(rspbody))

		entry.Debugf("Request Body: %s", requestBodyCopy)
		entry.Debugf("Request Headers: %s", req.Header)
		entry.Debugf("Response Body: %s", rspbody)
		entry.Debugf("Response Headers: %s", resp.Header)
	}
	return resp, nil
}
