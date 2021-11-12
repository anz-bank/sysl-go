package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/core"
	"google.golang.org/grpc"
)

var (
	errDefault = fmt.Errorf("")
)

func makeHeader(vals map[string]string) http.Header {
	head := make(http.Header)
	for k, v := range vals {
		head.Add(k, v)
	}

	return head
}

func equalFold(key string, actualVal, expectedVal []string) string {
	actualStr := strings.Join(actualVal, ", ")
	expectedStr := strings.Join(expectedVal, ", ")

	if !strings.EqualFold(expectedStr, actualStr) {
		return fmt.Sprintf("%s: '%s'!='%s'", key, actualStr, expectedStr)
	}

	return ""
}

// IgnoreHeaders is a list of headers that should be ignored during testing.
var IgnoreHeaders = []string{}

func shouldSkipTestingHeader(hdr string) bool {
	for _, v := range IgnoreHeaders {
		if strings.EqualFold(v, hdr) {
			return true
		}
	}

	return false
}

func iterateHeaders(actualHeader, expectedHeaders http.Header) (extra, missing, valMismatch []string) {
	for k, v := range actualHeader {
		if shouldSkipTestingHeader(k) {
			continue
		}
		expectedVal, ok := expectedHeaders[k]
		if !ok {
			extra = append(extra, k)

			continue
		}

		mismatch := equalFold(k, v, expectedVal)
		if mismatch != "" {
			valMismatch = append(valMismatch, mismatch)
		}
	}

	for k := range expectedHeaders {
		if shouldSkipTestingHeader(k) {
			continue
		}
		_, ok := actualHeader[k]
		if !ok {
			missing = append(missing, k)

			continue
		}
	}

	return extra, missing, valMismatch
}

func verifyHeaders(expected http.Header, actual http.Header, checkForExtra ...bool) error {
	extra, missing, valMismatch := iterateHeaders(actual, expected)

	errorStr := ""
	if len(checkForExtra) > 0 && checkForExtra[0] && extra != nil {
		errorStr += fmt.Sprintf("the following header fields were not expected: '%v'\n", extra)
	}
	if missing != nil {
		errorStr += fmt.Sprintf("the following header fields were expected but missing: '%v'\n", missing)
	}
	if valMismatch != nil {
		errorStr += fmt.Sprintf("the following header fields were received with incorrect values: '%v'\n", valMismatch)
	}
	if errorStr != "" {
		return fmt.Errorf("%s %w", errorStr, errDefault)
	}

	return nil
}

func expectHeadersExistImp(headers []string, actual http.Header) error {
	var missing []string
	for _, h := range headers {
		if _, exists := actual[http.CanonicalHeaderKey(h)]; !exists {
			missing = append(missing, h)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("Expected headers were missing: %s", missing)
	}

	return nil
}

func expectHeadersDoNotExistImp(headers []string, actual http.Header) error {
	var extra []string
	for _, h := range headers {
		if _, exists := actual[http.CanonicalHeaderKey(h)]; exists {
			extra = append(extra, h)
		}
	}

	if len(extra) > 0 {
		return fmt.Errorf("Headers were expected to be missing: %s", extra)
	}

	return nil
}

func expectHeadersExistExactlyImp(headers []string, actual http.Header) (missingError, extraError error) {
	var extra, missing []string
	m := map[string]interface{}{}

	for _, h := range headers {
		can := http.CanonicalHeaderKey(h)
		m[can] = nil
		if _, exists := actual[can]; !exists {
			missing = append(missing, h)
		}
	}
	for h := range actual {
		if _, exists := m[h]; !exists {
			extra = append(extra, h)
		}
	}

	if len(missing) > 0 {
		missingError = fmt.Errorf("Expected headers were missing: %s", missing)
	}
	if len(extra) > 0 {
		extraError = fmt.Errorf("Extra headers were found: %s", extra)
	}

	return missingError, extraError
}

func GetResponseBodyAndClose(b io.ReadCloser) []byte {
	if b == nil {
		return nil
	}
	var buf bytes.Buffer
	defer func() { _ = b.Close() }()
	if _, err := buf.ReadFrom(b); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func GetTestLine() string {
	for i := 1; i < 10; i++ {
		if _, file, line, ok := runtime.Caller(i); ok {
			base := filepath.Base(file)
			if strings.HasSuffix(base, "_test.go") {
				return fmt.Sprintf("Expectation At:  %s:%d", base, line)
			}
		} else {
			return ""
		}
	}

	return ""
}

// CreateServiceWithTestHooksPatched will return a function with the same signature as createService but will patch
// the test hooks into the result (we don't know the config type at this point so use reflection).
func CreateServiceWithTestHooksPatched(createService interface{}, testHooks *core.Hooks) interface{} {
	return reflect.MakeFunc(reflect.TypeOf(createService), func(args []reflect.Value) (results []reflect.Value) {
		// Call createService
		createServiceResult := reflect.ValueOf(createService).Call(args)
		if err := createServiceResult[2].Interface(); err != nil {
			return createServiceResult
		}

		// Patch in the test hooks
		h := createServiceResult[1].Interface().(*core.Hooks)
		if h == nil {
			h = testHooks
		} else {
			h.HTTPClientBuilder = testHooks.HTTPClientBuilder
			h.StoppableServerBuilder = testHooks.StoppableServerBuilder
			if testHooks.OverrideGrpcDialOptions != nil {
				if h.AdditionalGrpcDialOptions == nil {
					h.OverrideGrpcDialOptions = testHooks.OverrideGrpcDialOptions
				} else {
					additionalGrpcDialOptions := h.AdditionalGrpcDialOptions
					h.AdditionalGrpcDialOptions = nil

					h.OverrideGrpcDialOptions = func(serviceName string, cfg *config.CommonGRPCDownstreamData) ([]grpc.DialOption, error) {
						options, err := testHooks.OverrideGrpcDialOptions(serviceName, cfg)
						if err != nil {
							return nil, err
						}
						options = append(options, additionalGrpcDialOptions...)

						return options, nil
					}
				}
			}
			h.StoppableGrpcServerBuilder = testHooks.StoppableGrpcServerBuilder
			if h.ValidateConfig == nil {
				h.ValidateConfig = testHooks.ValidateConfig
			} else {
				h.ValidateConfig = func(ctx context.Context, cfg *config.DefaultConfig) error {
					_ = testHooks.ValidateConfig(ctx, cfg)

					return h.ValidateConfig(ctx, cfg)
				}
			}
			// Prefer the app set func over the test default
			if h.ShouldSetGrpcGlobalLogger == nil {
				h.ShouldSetGrpcGlobalLogger = testHooks.ShouldSetGrpcGlobalLogger
			}
		}
		createServiceResult[1] = reflect.ValueOf(h)

		return createServiceResult
	}).Interface()
}
