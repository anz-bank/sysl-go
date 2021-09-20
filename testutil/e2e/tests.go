package e2e

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/restlib"
	"github.com/stretchr/testify/assert"
)

type Tests func(t *testing.T, w http.ResponseWriter, r *http.Request)

type TestCall struct {
	Method       string
	URL          string
	Headers      map[string]string
	Body         string
	ExpectedCode int
	ExpectedBody string
	TestCodeFn   func(t *testing.T, expected, actual int)
	TestBodyFn   func(t *testing.T, expected, actual string)
}

type TestCall2 struct {
	Method       string
	URL          string
	Headers      map[string]string
	Body         []byte
	ExpectedCode *int
	ExpectedBody []byte
	TestCodeFn   func(t *testing.T, actual int)
	TestBodyFn   func(t *testing.T, actual []byte)
}

// ExpectHeaders: Expects the given headers and their values exist in the response.
// checkForExtra is an optional parameter to check for extra headers not expected.
func ExpectHeaders(headers map[string]string, checkForExtra ...bool) Tests {
	hdrs := makeHeader(headers)
	loc := GetTestLine()

	return func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		assert.NoError(t, verifyHeaders(hdrs, r.Header, checkForExtra...), loc)
	}
}

// ExpectHeadersExist: Expects the given header names can be found in the response.
func ExpectHeadersExist(headers []string) Tests {
	loc := GetTestLine()

	return func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		var missing []string
		for _, h := range headers {
			if _, exists := r.Header[http.CanonicalHeaderKey(h)]; !exists {
				missing = append(missing, h)
			}
		}

		if len(missing) > 0 {
			assert.Empty(t, missing, "Expected headers were missing. %s", loc)
		}
	}
}

// ExpectHeadersDoNotExist: Expects the given header names cannot be found in the response.
func ExpectHeadersDoNotExist(headers []string) Tests {
	loc := GetTestLine()

	return func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		var extra []string
		for _, h := range headers {
			if _, exists := r.Header[http.CanonicalHeaderKey(h)]; exists {
				extra = append(extra, h)
			}
		}

		if len(extra) > 0 {
			assert.Empty(t, extra, "Headers were expected to be missing. %s", loc)
		}
	}
}

// ExpectHeadersExistExactly: Expects the given headers in the response exist (and no others).
func ExpectHeadersExistExactly(headers []string) Tests {
	loc := GetTestLine()

	return func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		var extra, missing []string
		m := map[string]interface{}{}

		for _, h := range headers {
			can := http.CanonicalHeaderKey(h)
			m[can] = nil
			if _, exists := r.Header[can]; !exists {
				missing = append(missing, h)
			}
		}
		for h := range r.Header {
			if _, exists := m[h]; !exists {
				extra = append(extra, h)
			}
		}

		assert.Empty(t, missing, "Expected headers were missing. %s", loc)
		assert.Empty(t, extra, "Extra headers were found. %s", loc)
	}
}

func ExpectQueryParams(query map[string][]string) Tests {
	loc := GetTestLine()
	in := url.Values(query)

	return func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		expected := in.Encode()
		actual := r.URL.Query().Encode()
		assert.Equal(t, expected, actual, loc)
	}
}

func ExpectURLParam(key string, expected string) Tests {
	loc := GetTestLine()

	return func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		actual := restlib.GetURLParam(r, key)

		assert.Equal(t, expected, actual, loc)
	}
}

func ExpectURLParamForInt(key string, expected int64) Tests {
	loc := GetTestLine()

	return func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		actual := restlib.GetURLParamForInt(r, key)

		assert.Equal(t, expected, actual, loc)
	}
}

func ExpectBody(expected []byte) Tests {
	loc := GetTestLine()

	return func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		body := GetResponseBodyAndClose(r.Body)
		assert.Equal(t, expected, body, loc)
	}
}

func ExpectJSONBody(expected []byte) Tests {
	loc := GetTestLine()

	return func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		body := GetResponseBodyAndClose(r.Body)
		assert.JSONEq(t, string(expected), string(body), loc)
	}
}

func Response(code int, headers map[string]string, body []byte) Tests {
	hdrs := map[string]string{}
	for k, v := range headers {
		hdrs[k] = v
	}
	bdy := make([]byte, len(body))
	_ = copy(bdy, body)

	return func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		for k, v := range hdrs {
			w.Header().Set(k, v)
		}
		w.WriteHeader(code)
		_, _ = w.Write(bdy)
	}
}

func ForceDownstreamTimeout() Tests {
	return func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		<-time.After(DownstreamTimeout + 100*time.Millisecond)
	}
}
