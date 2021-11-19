package e2e

import (
	"net/http"
	"net/url"
	"time"

	"github.com/anz-bank/sysl-go/restlib"
	"github.com/anz-bank/sysl-go/syslgo"
	"github.com/stretchr/testify/assert"
)

type Tests func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request)
type ResponseTest func(t syslgo.TestingT, actual *http.Response)

type TestCall struct {
	Method       string
	URL          string
	Headers      map[string]string
	Body         string
	ExpectedCode int
	ExpectedBody string
	TestCodeFn   func(t syslgo.TestingT, expected, actual int)
	TestBodyFn   func(t syslgo.TestingT, expected, actual string)
}

type TestCall2 struct {
	Method       string
	URL          string
	Headers      map[string]string
	Body         []byte
	ExpectedCode *int
	ExpectedBody []byte
	TestCodeFn   func(t syslgo.TestingT, actual int)
	TestBodyFn   func(t syslgo.TestingT, actual []byte)
	TestRespFns  []ResponseTest
}

// ExpectResponseHeaders: Expects the given headers and their values exist in the response.
// checkForExtra is an optional parameter to check for extra headers not expected.
func ExpectResponseHeaders(headers map[string]string, checkForExtra ...bool) ResponseTest {
	hdrs := makeHeader(headers)
	loc := GetTestLine()

	return func(t syslgo.TestingT, actual *http.Response) {
		assert.NoError(t, verifyHeaders(hdrs, actual.Header, checkForExtra...), loc)
	}
}

// ExpectResponseHeadersExist: Expects the given header names can be found in the response.
func ExpectResponseHeadersExist(headers []string) ResponseTest {
	loc := GetTestLine()

	return func(t syslgo.TestingT, actual *http.Response) {
		assert.NoError(t, expectHeadersExistImp(headers, actual.Header), loc)
	}
}

// ExpectResponseHeadersDoNotExist: Expects the given header names cannot be found in the response.
func ExpectResponseHeadersDoNotExist(headers []string) ResponseTest {
	loc := GetTestLine()

	return func(t syslgo.TestingT, actual *http.Response) {
		assert.NoError(t, expectHeadersDoNotExistImp(headers, actual.Header), loc)
	}
}

// ExpectResponseHeadersExistExactly: Expects the given headers in the response exist (and no others).
func ExpectResponseHeadersExistExactly(headers []string) ResponseTest {
	loc := GetTestLine()

	return func(t syslgo.TestingT, actual *http.Response) {
		missingError, extraError := expectHeadersExistExactlyImp(headers, actual.Header)
		assert.NoError(t, missingError, loc)
		assert.NoError(t, extraError, loc)
	}
}

// ExpectHeaders: Expects the given headers and their values exist in the response.
// checkForExtra is an optional parameter to check for extra headers not expected.
func ExpectHeaders(headers map[string]string, checkForExtra ...bool) Tests {
	hdrs := makeHeader(headers)
	loc := GetTestLine()

	return func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
		assert.NoError(t, verifyHeaders(hdrs, r.Header, checkForExtra...), loc)
	}
}

// ExpectHeadersExist: Expects the given header names can be found in the response.
func ExpectHeadersExist(headers []string) Tests {
	loc := GetTestLine()

	return func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
		assert.NoError(t, expectHeadersExistImp(headers, r.Header), loc)
	}
}

// ExpectHeadersDoNotExist: Expects the given header names cannot be found in the response.
func ExpectHeadersDoNotExist(headers []string) Tests {
	loc := GetTestLine()

	return func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
		assert.NoError(t, expectHeadersDoNotExistImp(headers, r.Header), loc)
	}
}

// ExpectHeadersExistExactly: Expects the given headers in the response exist (and no others).
func ExpectHeadersExistExactly(headers []string) Tests {
	loc := GetTestLine()

	return func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
		missingError, extraError := expectHeadersExistExactlyImp(headers, r.Header)
		assert.NoError(t, missingError, loc)
		assert.NoError(t, extraError, loc)
	}
}

func ExpectQueryParams(query map[string][]string) Tests {
	loc := GetTestLine()
	in := url.Values(query)

	return func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
		expected := in.Encode()
		actual := r.URL.Query().Encode()
		assert.Equal(t, expected, actual, loc)
	}
}

func ExpectURLParam(key string, expected string) Tests {
	loc := GetTestLine()

	return func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
		actual := restlib.GetURLParam(r, key)

		assert.Equal(t, expected, actual, loc)
	}
}

func ExpectURLParamForInt(key string, expected int64) Tests {
	loc := GetTestLine()

	return func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
		actual := restlib.GetURLParamForInt(r, key)

		assert.Equal(t, expected, actual, loc)
	}
}

func ExpectBody(expected []byte) Tests {
	loc := GetTestLine()

	return func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
		body := GetResponseBodyAndClose(r.Body)
		assert.Equal(t, expected, body, loc)
	}
}

func ExpectJSONBody(expected []byte) Tests {
	loc := GetTestLine()

	return func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
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

	return func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
		for k, v := range hdrs {
			w.Header().Set(k, v)
		}
		w.WriteHeader(code)
		_, _ = w.Write(bdy)
	}
}

func ForceDownstreamTimeout() Tests {
	return func(t syslgo.TestingT, w http.ResponseWriter, r *http.Request) {
		<-time.After(DownstreamTimeout + 100*time.Millisecond)
	}
}
