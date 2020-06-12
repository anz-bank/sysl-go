package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/anz-bank/pkg/log"
)

type unclosedResponseBodyMonitorContextKey struct{}

type unclosedResponseBodyMonitor struct {
	mtx           *sync.Mutex
	openResponses []*readCloserClosedWrapper
}

func AddResponseBodyMonitorToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, unclosedResponseBodyMonitorContextKey{}, newUnclosedResponseBodyMonitor())
}

func AddResponseToMonitor(ctx context.Context, resp *http.Response) {
	if val, ok := ctx.Value(unclosedResponseBodyMonitorContextKey{}).(*unclosedResponseBodyMonitor); ok {
		val.addResponse(resp)
	}
}

func CheckForUnclosedResponses(ctx context.Context) {
	if val, ok := ctx.Value(unclosedResponseBodyMonitorContextKey{}).(*unclosedResponseBodyMonitor); ok {
		openBodyErrors := OpenResponseBodyErrors{}
		for _, body := range val.getResponsesWithOpenBodies() {
			err := openBodyError{
				cause: body.parentReq.URL.String(),
				err:   "response body not closed",
			}
			openBodyErrors.errors = append(openBodyErrors.errors, err)
		}

		openBodyCount := len(openBodyErrors.errors)
		if openBodyCount > 0 {
			err := errors.New(openBodyErrors.Error())
			log.Error(ctx, err)
			panic(err)
		}
	}
}

func newUnclosedResponseBodyMonitor() *unclosedResponseBodyMonitor {
	return &unclosedResponseBodyMonitor{
		mtx:           &sync.Mutex{},
		openResponses: []*readCloserClosedWrapper{},
	}
}

func (r *unclosedResponseBodyMonitor) addResponse(rsp *http.Response) {
	wrapper := &readCloserClosedWrapper{
		rsp.Request,
		rsp.Body,
		false,
	}
	rsp.Body = wrapper
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.openResponses = append(r.openResponses, wrapper)
}

func (r *unclosedResponseBodyMonitor) getResponsesWithOpenBodies() []*readCloserClosedWrapper {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	openResponses := make([]*readCloserClosedWrapper, 0)
	for _, val := range r.openResponses {
		if !val.closed {
			openResponses = append(openResponses, val)
		}
	}
	return openResponses
}

type readCloserClosedWrapper struct {
	parentReq *http.Request
	parent    io.ReadCloser
	closed    bool
}

type openBodyError struct {
	cause string
	err   string
}

type OpenResponseBodyErrors struct {
	errors []openBodyError
}

func (e *OpenResponseBodyErrors) Error() string {
	var errors string
	for _, err := range e.errors {
		errors += fmt.Sprintf("%#v \n", err)
	}
	return fmt.Sprintf("%#v", errors)
}

func (r *readCloserClosedWrapper) Read(p []byte) (n int, err error) {
	return r.parent.Read(p)
}
func (r *readCloserClosedWrapper) Close() error {
	r.closed = true
	return r.parent.Close()
}
