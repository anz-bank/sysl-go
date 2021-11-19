package e2e

import (
	"net/http"
	"strings"
	"sync"

	"github.com/anz-bank/sysl-go/syslgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type restOrderedHandlers struct {
	sync.Mutex
	t        syslgo.TestingT
	handlers []http.HandlerFunc
}

func (o *restOrderedHandlers) add(h http.HandlerFunc) {
	o.Lock()
	defer o.Unlock()

	o.handlers = append(o.handlers, h)
}

const failedTestStatusCode = 599

func (o *restOrderedHandlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	o.Lock()
	defer o.Unlock()
	if len(o.handlers) == 0 {
		o.t.Log("Unexpected downstream call")
		w.WriteHeader(failedTestStatusCode)
		_, _ = w.Write([]byte("TEST FAILED"))

		return
	}
	o.handlers[0].ServeHTTP(w, r)
	o.handlers = o.handlers[1:]
}

func (o *restOrderedHandlers) assertCompleted(hostname, methodAndPath string) {
	o.Lock()
	defer o.Unlock()

	assert.Empty(o.t, o.handlers, "Endpoint '%s -> %s' has %d un-hit expected calls",
		hostname, methodAndPath, len(o.handlers))
}

func (o *restOrderedHandlers) Expect(tests ...Tests) Endpoint {
	require.NotEmpty(o.t, tests, "No tests supplied")
	o.add(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if b := r.Body; b != nil {
				_ = b.Close() // we don't care about errors here
			}
			if r := recover(); r != nil {
				o.t.Fatal(r)
			}
		}()
		for _, testFn := range tests {
			testFn(o.t, w, r)
		}
	})

	return o
}

const (
	ContentTypeKey  = "Content-Type"
	numRegularTests = 3
)

func (o *restOrderedHandlers) ExpectSimple(expectedHeaders map[string]string, expectedBody []byte, returnCode int,
	returnHeaders map[string]string, returnBody []byte, extraTests ...Tests) Endpoint {
	tests := make([]Tests, 0, len(extraTests)+numRegularTests)
	tests = append(tests, ExpectHeaders(expectedHeaders))

	if ct, ok := expectedHeaders[ContentTypeKey]; ok && strings.Contains(ct, "application/json") && len(expectedBody) != 0 {
		tests = append(tests, ExpectJSONBody(expectedBody))
	} else {
		tests = append(tests, ExpectBody(expectedBody))
	}

	tests = append(tests, extraTests...)
	tests = append(tests, Response(returnCode, returnHeaders, returnBody))

	return o.Expect(tests...)
}
