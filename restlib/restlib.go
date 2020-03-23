package restlib

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/anz-bank/sysl-go/common"
	"github.com/pkg/errors"
)

// HTTPResult is the result return by the library
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
		return makeHTTPResult(resp, body, nil), nil
	}

	e := reflect.ValueOf(respStruct).Elem()
	if e.Kind() == reflect.String {
		p := reflect.New(e.Type())
		p.Elem().Set(reflect.ValueOf(body).Convert(e.Type()))
		return makeHTTPResult(resp, body, p.Interface()), nil
	}

	contentType := resp.Header.Get("Content-Type")
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

// DoHTTPRequest returns HTTPResult
//nolint:funlen // TODO: Refactor this function to be shorter
func DoHTTPRequest(ctx context.Context, client *http.Client, method string,
	urlString string, body interface{}, required []string,
	okResponse interface{}, errorResponse interface{}) (*HTTPResult, error) {
	var reader io.Reader
	headers := common.RequestHeaderFromContext(ctx)
	contentType := headers.Get("Content-Type")

	// Validations 1:
	// If we have body, marshal it to json
	if body != nil {
		if strings.Contains(contentType, "xml") {
			var strBody string
			strBody = reflect.ValueOf(body).Convert(reflect.TypeOf(strBody)).String()
			if strings.HasSuffix(strBody, " Value>") {
				return nil, errors.Errorf(`Incompatible type as xml body: %s`, strBody)
			}
			reader = strings.NewReader(strBody)
		} else {
			reqJSON, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			reader = bytes.NewReader(reqJSON)
		}
	}

	// Validations 2:
	// if we have required headers, see if they have been passed to us
	for _, key := range required {
		if has := headers.Get(key); has == "" {
			return nil, errors.Errorf("Missing Required header: %s", key)
		}
	}

	httpRequest, err := http.NewRequest(method, urlString, reader)
	if err != nil {
		return nil, err
	}

	httpRequest.Header = headers

	httpResponse, err := client.Do(httpRequest.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer httpResponse.Body.Close()

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
	if httpResponse.StatusCode == http.StatusOK ||
		httpResponse.StatusCode == http.StatusCreated ||
		httpResponse.StatusCode == http.StatusAccepted {
		return unmarshal(httpResponse, respBody, okResponse)
	}

	// Error
	result, err := unmarshal(httpResponse, respBody, errorResponse)
	if err != nil {
		return nil, err
	}

	// Successful unmarshal but we have unmarshalled an error.
	return nil, result
}

// SendHTTPResponse sends the http response to the client
func SendHTTPResponse(w http.ResponseWriter, httpStatus int, responses ...interface{}) {
	w.WriteHeader(httpStatus)

	for _, resp := range responses {
		if resp != nil {
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
	}
}

// SetHeaders sets the headers in response
func SetHeaders(w http.ResponseWriter, headerMap http.Header) {
	for k, v := range headerMap {
		for _, hv := range v {
			w.Header().Add(k, hv)
		}
	}
}
