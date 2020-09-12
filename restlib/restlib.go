package restlib

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/anz-bank/sysl-go/common"
	"github.com/pkg/errors"
)

// HTTPResult is the result return by the library.
type HTTPResult struct {
	HTTPResponse *http.Response
	Body         []byte
	Response     interface{}
}

func (r *HTTPResult) Error() string {
	return r.HTTPResponse.Status
}

func makeHTTPResult(res *http.Response, body []byte, resp interface{}) *HTTPResult {
	return &HTTPResult{
		HTTPResponse: res,
		Body:         body,
		Response:     resp,
	}
}

func unmarshal(resp *http.Response, body []byte, respStruct interface{}) (*HTTPResult, error) {
	if resp == nil {
		panic("unmarshal expecting a non-nil http.Response")
	}
	if respStruct == nil || body == nil || len(body) == 0 {
		// Obtain the respStruct's dynamic type and check if its a pointer
		if p := reflect.TypeOf(respStruct); p != nil && p.Kind() == reflect.Ptr {
			// Dereference the dynamic pointer type and pass the created zero value
			return makeHTTPResult(resp, body, reflect.New(p.Elem()).Interface()), nil
		}
		return makeHTTPResult(resp, body, nil), nil
	}

	contentType := resp.Header.Get("Content-Type")

	e := reflect.ValueOf(respStruct).Elem()
	kind := e.Kind()

	if kind == reflect.String || (kind == reflect.Slice && e.Type().Elem().Name() == "uint8") {
		p := reflect.New(e.Type())
		p.Elem().Set(reflect.ValueOf(body).Convert(e.Type()))
		return makeHTTPResult(resp, body, p.Interface()), nil
	}

	if strings.Contains(contentType, "xml") {
		respStruct = string(body)
	} else {
		err := json.Unmarshal(body, respStruct)
		if err != nil {
			return makeHTTPResult(resp, body, nil), err
		}
	}
	return makeHTTPResult(resp, body, respStruct), nil
}

// DoHTTPRequest returns HTTPResult.
func DoHTTPRequest(ctx context.Context,
	client *http.Client,
	method, urlString string,
	body interface{},
	required []string,
	okResponse, errorResponse interface{},
) (*HTTPResult, error) {
	return DoHTTPRequest2(ctx, &HTTPRequest{
		client,
		method,
		urlString,
		body,
		required,
		okResponse,
		errorResponse,
		nil,
	})
}

type HTTPRequest struct {
	Client                    *http.Client
	Method, URLString         string
	Body                      interface{}
	Required                  []string
	OKResponse, ErrorResponse interface{}
	ExtraHeaders              map[string][]string
}

// DoHTTPRequest returns HTTPResult.
//nolint:funlen // TODO: Refactor this function to be shorter.
func DoHTTPRequest2(ctx context.Context, config *HTTPRequest) (*HTTPResult, error) {
	var reader io.Reader
	headers := common.RequestHeaderFromContext(ctx)
	contentType := headers.Get("Content-Type")

	// Validations 1:
	// If we have body, marshal it to json
	if config.Body != nil {
		if strings.Contains(contentType, "xml") {
			var strBody string
			strBody = reflect.ValueOf(config.Body).Convert(reflect.TypeOf(strBody)).String()
			if strings.HasSuffix(strBody, " Value>") {
				return nil, errors.Errorf(`Incompatible type as xml body: %s`, strBody)
			}
			reader = strings.NewReader(strBody)
		} else {
			reqJSON, err := json.Marshal(config.Body)
			if err != nil {
				return nil, err
			}
			reader = bytes.NewReader(reqJSON)
		}
	}

	if config.ExtraHeaders != nil {
		for key, values := range config.ExtraHeaders {
			for _, v := range values {
				headers.Add(key, v)
			}
		}
	}

	// Validations 2:
	// if we have required headers, see if they have been passed to us
	for _, key := range config.Required {
		if has := headers.Get(key); has == "" {
			return nil, errors.Errorf("Missing Required header: %s", key)
		}
	}

	httpRequest, err := http.NewRequest(config.Method, config.URLString, reader)
	if err != nil {
		return nil, err
	}

	httpRequest.Header = headers

	httpResponse, err := config.Client.Do(httpRequest.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer httpResponse.Body.Close()

	// TODO: remove this after confirming that gzip was handled by lower layer
	var bodyReader io.Reader
	if m, _ := regexp.MatchString(`(?i)gzip`, httpResponse.Header.Get("Content-Encoding")); m {
		bodyReader, err = gzip.NewReader(httpResponse.Body)
		if err != nil {
			return nil, err
		}
	} else {
		bodyReader = httpResponse.Body
	}

	respBody, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}

	// OK
	switch httpResponse.StatusCode {
	case http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNonAuthoritativeInfo,
		http.StatusNoContent,
		http.StatusResetContent:
		return unmarshal(httpResponse, respBody, config.OKResponse)
	}

	// Error
	result, err := unmarshal(httpResponse, respBody, config.ErrorResponse)
	if err != nil {
		return nil, err
	}

	// Successful unmarshal but we have unmarshalled an error.
	return nil, result
}

// SendHTTPResponse sends the http response to the client.
func SendHTTPResponse(w http.ResponseWriter, httpStatus int, responses ...interface{}) {
	w.WriteHeader(httpStatus)

	contentType := w.Header().Get("Content-Type")

	for _, resp := range responses {
		if resp != nil {
			switch {
			case strings.Contains(contentType, "xml"):
				_ = xml.NewEncoder(w).Encode(resp)
			case strings.Contains(contentType, "image"):
				_, _ = w.Write(reflect.ValueOf(resp).Elem().Bytes())
			case strings.Contains(contentType, "octet-stream"), strings.Contains(contentType, "pdf"):
				switch data := resp.(type) {
				case *[]byte:
					_, _ = w.Write(*data)
				case []byte:
					_, _ = w.Write(data)
				}
			default:
				_ = json.NewEncoder(w).Encode(resp)
			}
			return
		}
	}
}

// SetHeaders sets the headers in response.
func SetHeaders(w http.ResponseWriter, headerMap http.Header) {
	for k, v := range headerMap {
		for _, hv := range v {
			w.Header().Add(k, hv)
		}
	}
}

// OnRestResultHTTPResult is called from generated code when an HTTP result is retrieved.
// The current implementation of restlib.DoHTTPRequest returns an *HTTPResult as an error when a non-
// successful status code is received. The implementation of this method relies on this behaviour.
// to set the rest result in the event of a failed request.
func OnRestResultHTTPResult(ctx context.Context, result *HTTPResult, err error) {
	if result != nil {
		SetRestResult(ctx, toRestResult(*result))
	} else if res, ok := err.(*HTTPResult); ok {
		SetRestResult(ctx, toRestResult(*res))
	}
}

func toRestResult(result HTTPResult) common.RestResult {
	return common.RestResult{
		StatusCode: result.HTTPResponse.StatusCode,
		Headers:    result.HTTPResponse.Header,
		Body:       result.Body,
	}
}

// SetRestResult the contents of the common.RestResult stored in the context.
// The RestResult is stored in the context through the common.ProvisionRestResult method.
// This method is exported so that unit tests can set the rest result with appropriate
// values as required.
func SetRestResult(ctx context.Context, result common.RestResult) {
	raw := ctx.Value(common.RestResultContextKey{})
	if raw == nil {
		return
	}
	target := raw.(*common.RestResult)
	target.Body = result.Body
	target.Headers = result.Headers
	target.StatusCode = result.StatusCode
}
