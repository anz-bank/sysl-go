package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/anz-bank/pkg/log"
)

// Timeout is a middleware that cancels ctx after a given timeout or call a special handler on timeout.
func Timeout(ctx context.Context, timeout time.Duration, timeoutHandler http.Handler) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return TimeoutHandler(ctx, next, timeout, timeoutHandler)
	}
}

// Forked from go/src/net/http/server.go
// Changes:
// * Accept a http.Handler instead of a string for the error message handling
// * Log panics to logrus instead of re-panic()-ing with a new stack trace

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TimeoutHandler returns a Handler that runs h with the given time limit.
//
// The new Handler calls h.ServeHTTP to handle each request, but if a
// call runs for longer than its time limit, the handler calls the error handler
// which can then return a HTTP result however is deems correct.
// After such a timeout, writes by h to its ResponseWriter will return
// ErrHandlerTimeout.
//
// TimeoutHandler buffers all Handler writes to memory and does not
// support the Hijacker or Flusher interfaces.
func TimeoutHandler(ctx context.Context, h http.Handler, dt time.Duration, timeout http.Handler) http.Handler {
	return &timeoutHandler{
		ctx:            ctx,
		handler:        h,
		timeoutHandler: timeout,
		dt:             dt,
	}
}

// ErrHandlerTimeout is returned on ResponseWriter Write calls
// in handlers which have timed out.
var ErrHandlerTimeout = errors.New("http: Handler timeout")

type timeoutHandler struct {
	ctx            context.Context
	handler        http.Handler
	timeoutHandler http.Handler
	dt             time.Duration

	// When set, no context will be created and this context will
	// be used instead.
	testContext context.Context
}

func (h *timeoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := h.testContext
	if ctx == nil {
		var cancelCtx context.CancelFunc
		ctx, cancelCtx = context.WithTimeout(r.Context(), h.dt)
		defer cancelCtx()
	}
	r = r.WithContext(ctx)
	done := make(chan struct{})
	tw := &timeoutWriter{
		w: w,
		h: make(http.Header),
	}
	panicChan := make(chan interface{}, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				log.Error(ctx, errors.New(string(debug.Stack())))
				panicChan <- p
			}
		}()
		h.handler.ServeHTTP(tw, r)
		close(done)
	}()
	timeoutFunc := func() {
		tw.mu.Lock()
		defer tw.mu.Unlock()
		h.timeoutHandler.ServeHTTP(w, r)
		tw.timedOut = true
	}
	select {
	case <-done:
		tw.mu.Lock()
		defer tw.mu.Unlock()
		dst := w.Header()
		for k, vv := range tw.h {
			dst[k] = vv
		}
		if !tw.wroteHeader {
			tw.code = http.StatusOK
		}
		w.WriteHeader(tw.code)
		_, _ = w.Write(tw.wbuf.Bytes())
	case <-panicChan:
		timeoutFunc()
	case <-ctx.Done():
		timeoutFunc()
	}
}

type timeoutWriter struct {
	w    http.ResponseWriter
	h    http.Header
	wbuf bytes.Buffer

	mu          sync.Mutex
	timedOut    bool
	wroteHeader bool
	code        int
}

func (tw *timeoutWriter) Header() http.Header { return tw.h }

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return 0, ErrHandlerTimeout
	}
	if !tw.wroteHeader {
		tw.writeHeader(http.StatusOK)
	}
	return tw.wbuf.Write(p)
}

func (tw *timeoutWriter) WriteHeader(code int) {
	checkWriteHeaderCode(code)
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut || tw.wroteHeader {
		return
	}
	tw.writeHeader(code)
}

func (tw *timeoutWriter) writeHeader(code int) {
	tw.wroteHeader = true
	tw.code = code
}

func checkWriteHeaderCode(code int) {
	// Issue 22880: require valid WriteHeader status codes.
	// For now we only enforce that it's three digits.
	// In the future we might block things over 599 (600 and above aren't defined
	// at https://httpwg.org/specs/rfc7231.html#status.codes)
	// and we might block under 200 (once we have more mature 1xx support).
	// But for now any three digits.
	//
	// We used to send "HTTP/1.1 000 0" on the wire in responses but there's
	// no equivalent bogus thing we can realistically send in HTTP/2,
	// so we'll consistently panic instead and help people find their bugs
	// early. (We can't return an error from WriteHeader even if we wanted to.)
	if code < 100 || code > 999 {
		panic(fmt.Sprintf("invalid WriteHeader code %v", code))
	}
}
