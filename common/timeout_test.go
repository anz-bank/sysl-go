package common

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testHandler struct {
	called   bool
	httpCode int
	body     []byte
}

func defaultTestHandler() *testHandler {
	return &testHandler{
		false,
		500,
		[]byte("hello"),
	}
}

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.called = true
	w.WriteHeader(t.httpCode)
	_, _ = w.Write(t.body)
}

func TestTimeoutHandler_NoCallbackCalledIfNotTimeout(t *testing.T) {
	req := require.New(t)
	tester := defaultTestHandler()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200); _, _ = w.Write([]byte("OK")) })

	timeoutmware := Timeout(context.Background(), time.Second, tester)
	ts := httptest.NewServer(timeoutmware(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	req.NoError(err)
	body, err := ioutil.ReadAll(resp.Body)
	req.NoError(err)
	req.Equal("OK", string(body))
	req.Equal(200, resp.StatusCode)
	req.False(tester.called)
	defer resp.Body.Close()
}

func TestTimeoutHandler_CallbackCalledIfTimeout(t *testing.T) {
	req := require.New(t)
	tester := defaultTestHandler()

	doneChan := make(chan bool)
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-doneChan
		_, _ = w.Write([]byte("OK"))
	})

	timeoutmware := Timeout(context.Background(), time.Millisecond, tester)

	ts := httptest.NewServer(timeoutmware(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	doneChan <- true
	req.NoError(err)
	body, err := ioutil.ReadAll(resp.Body)
	req.NoError(err)
	req.Equal("hello", string(body))
	req.Equal(500, resp.StatusCode)
	req.True(tester.called)

	defer resp.Body.Close()
}

func TestTimeoutHandler_NoPanicRethrow(t *testing.T) {
	req := require.New(t)
	tester := defaultTestHandler()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		panic("HElp")
	})

	timeoutmware := Timeout(context.Background(), time.Millisecond, tester)
	ts := httptest.NewServer(timeoutmware(handler))
	defer ts.Close()

	req.NotPanics(func() {
		resp, err := http.Get(ts.URL)
		req.NoError(err)
		body, err := ioutil.ReadAll(resp.Body)
		req.NoError(err)
		req.Equal("hello", string(body))
		req.Equal(500, resp.StatusCode)
		req.True(tester.called)
		defer resp.Body.Close()
	})
}

func TestTimeoutHandler_ContextTimoutMoreThanWriteTimeout(t *testing.T) {
	req := require.New(t)
	tester := defaultTestHandler()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(20 * time.Millisecond):
			w.WriteHeader(504)
			_, _ = w.Write([]byte("Timeout"))
		case <-r.Context().Done():
			return
		}
	})

	timeoutmware := Timeout(context.Background(), 10*time.Millisecond, tester)
	ts := httptest.NewUnstartedServer(timeoutmware(handler))
	ts.Config.WriteTimeout = 5 * time.Millisecond
	ts.Start()
	//nolint:bodyclose // We don't check the body
	_, err := http.Get(ts.URL)
	req.Equal(err.(*url.Error).Err.Error(), "EOF")
}

func TestTimeoutHandler_ContextTimoutLessThanWriteTimeout(t *testing.T) {
	req := require.New(t)
	tester := defaultTestHandler()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(20 * time.Millisecond):
			w.WriteHeader(504)
			_, _ = w.Write([]byte("Timeout"))
		case <-r.Context().Done():
			return
		}
	})

	timeoutmware := Timeout(context.Background(), 5*time.Millisecond, tester)
	ts := httptest.NewUnstartedServer(timeoutmware(handler))
	ts.Config.WriteTimeout = 10 * time.Millisecond
	ts.Start()

	resp, err := http.Get(ts.URL)
	req.NoError(err)
	body, err := ioutil.ReadAll(resp.Body)
	req.NoError(err)
	req.Equal(500, resp.StatusCode)
	req.Equal("hello", string(body))
	req.True(tester.called)
	defer resp.Body.Close()
}

func TestTimeoutHandler_ContextTimoutAndWriteTimeoutTooShort(t *testing.T) {
	req := require.New(t)
	tester := defaultTestHandler()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(5 * time.Millisecond):
			w.WriteHeader(504)
			_, _ = w.Write([]byte("Timeout"))
		case <-r.Context().Done():
			return
		}
	})

	timeoutmware := Timeout(context.Background(), 10*time.Millisecond, tester)
	ts := httptest.NewUnstartedServer(timeoutmware(handler))
	ts.Config.WriteTimeout = 10 * time.Millisecond
	ts.Start()

	resp, err := http.Get(ts.URL)
	req.NoError(err)
	body, err := ioutil.ReadAll(resp.Body)
	req.NoError(err)
	req.Equal("Timeout", string(body))
	req.Equal(504, resp.StatusCode)
	req.False(tester.called)
	defer resp.Body.Close()
}
