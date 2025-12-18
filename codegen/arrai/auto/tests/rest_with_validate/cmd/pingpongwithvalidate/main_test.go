package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/core"
	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/require"

	"rest_with_validate/internal/gen/pkg/servers/pingpongwithvalidate"
)

const applicationConfig = `---
genCode:
  upstream:
    http:
      basePath: "/"
      common:
        hostName: "localhost"
        port: 9021 # FIXME no guarantee this port is free
  downstream:
    contextTimeout: "30s"
`

type ResponseError struct {
	StatusCode int
	Body       []byte
}

func (r *ResponseError) Error() string {
	return fmt.Sprintf("code: %d, body: %s", r.StatusCode, r.Body)
}

func doPongPongRequestResponse(ctx context.Context, identifier int64, value int64) (int, int, error) {
	url := fmt.Sprintf("http://localhost:9021/pong-pong")

	return doPingPongRequestResponseimpl(ctx, url, identifier, value)
}

func doPingPongRequestResponseimpl(ctx context.Context, url string, identifier int64, value int64) (int, int, error) {
	type payload struct {
		Identifier *int64 `json:"identifier,omitempty"`
		Value      *int64 `json:"value,omitempty"`
	}

	requestObj := payload{
		Identifier: &identifier,
		Value:      &value,
	}
	if identifier == -1 {
		requestObj.Identifier = nil
	}
	if value == -1 {
		requestObj.Value = nil
	}

	requestData, err := json.Marshal(&requestObj)
	if err != nil {
		return -1, -1, err
	}

	return doPingRequestResponseImpl(ctx, "POST", url, bytes.NewReader(requestData))
}

func doPingRequestResponseImpl(ctx context.Context, method string, url string, body io.Reader) (int, int, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return -1, -1, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return -1, -1, err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, -1, err
	}
	if resp.StatusCode != 200 {
		return -1, -1, &ResponseError{resp.StatusCode, data}
	}
	var obj struct {
		Identifier int `json:"identifier"`
		Value      int `json:"value"`
	}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return -1, -1, err
	}
	return obj.Identifier, obj.Value, nil
}

func TestApplicationSmokeTest(t *testing.T) {
	// Override sysl-go app command line interface to directly pass in app config
	ctx := core.WithConfigFile(context.Background(), []byte(applicationConfig))

	appServer, err := newAppServer(ctx)
	require.NoError(t, err)
	defer func() {
		err := appServer.Stop()
		if err != nil {
			panic(err)
		}
	}()

	// Start application server
	go func() {
		err := appServer.Start()
		if err != nil {
			panic(err)
		}
	}()

	// Wait for application to come up
	backoff := retry.NewFibonacci(10 * time.Millisecond)
	backoff = retry.WithMaxDuration(10*time.Second, backoff)
	err = retry.Do(ctx, backoff, func(ctx context.Context) error {
		_, _, err := doPongPongRequestResponse(ctx, 0, 0)
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	// Test various combinations of request data for the pong-pong endpoint.
	// The request is not validated for missing parameters

	// Test a successful request
	identifier, value, err := doPongPongRequestResponse(ctx, 1, 1)
	require.Nil(t, err)
	require.Equal(t, 1, identifier)
	require.Equal(t, 1, value)

	// Test a request that fails due to a missing request parameter (identifier)
	identifier, value, err = doPongPongRequestResponse(ctx, -1, 1)
	require.Equal(t, 400, err.(*ResponseError).StatusCode)

	// Test a request that fails due to a missing request parameter (value)
	identifier, value, err = doPongPongRequestResponse(ctx, 1, -1)
	require.Equal(t, 400, err.(*ResponseError).StatusCode)
}

func getPopulatedBody() pingpongwithvalidate.PingWithValidateRequest {
	return pingpongwithvalidate.PingWithValidateRequest{
		ValidLength: common.NewString("12"),
		ValidSize:   3,
	}
}

func getPopulatedHeader() map[string]string {
	return map[string]string{
		"headerPattern": "1.1.1.1",
		"headerLength":  "aaa",
	}
}

type varToSet int

const (
	validLength varToSet = iota
	validLengthNil
	validSize
	exclusiveSize
	exclusiveSizeOld
	nonExclusiveSizeOld
	patternSimple
	patternWithNegativeLookahead
	enumString
	enumStringRef
	enumInt
	largeNumber
	arrayOfObject
	arrayOfEnum
	floatWithFractional
	refOfArrayWithLength
)

func getBody(vts varToSet, s string, i int64) pingpongwithvalidate.PingWithValidateRequest {
	ret := getPopulatedBody()

	switch vts {
	case validLength:
		ret.ValidLength = &s
	case validLengthNil:
		ret.ValidLength = nil
	case validSize:
		ret.ValidSize = i
	case exclusiveSize:
		ret.ExclusiveSize = &i
	case exclusiveSizeOld:
		ret.ExclusiveSizeOld = &i
	case nonExclusiveSizeOld:
		ret.NonExclusiveSizeOld = &i
	case patternSimple:
		ret.PatternSimple = &s
	case patternWithNegativeLookahead:
		ret.PatternWithNegativeLookahead = &s
	case enumString:
		ret.EnumString = &s
	case enumStringRef:
		ret.EnumStringRef = &s
	case enumInt:
		ret.EnumInt = &i
	case largeNumber:
		ret.LargeNumber = &i
	case arrayOfObject:
		ret.ArrayOfObjects = []pingpongwithvalidate.ArrayObjectDetails{{&s}}
	case arrayOfEnum:
		ret.ArrayOfEnumString = []string{s}
	case refOfArrayWithLength:
		ret.RefOfArrayWithLength = []string{s}
	}

	return ret
}

func getBodyFloat(vts varToSet, f float64) pingpongwithvalidate.PingWithValidateRequest {
	ret := getPopulatedBody()

	switch vts {
	case floatWithFractional:
		ret.FloatWithFractional = &f
	}

	return ret
}

func TestValidate_BodyParams(t *testing.T) {
	t.Parallel()
	gatewayTester := pingpongwithvalidate.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	for _, test := range []struct {
		name string
		code int
		body pingpongwithvalidate.PingWithValidateRequest
	}{
		{`allPopulated`, 200, getPopulatedBody()},

		{`missingOptional`, 200, getBody(validLengthNil, "", 0)},

		{`LengthTooShort`, 400, getBody(validLength, "1", 0)},
		{`LengthSmallest`, 200, getBody(validLength, "12", 0)},
		{`LengthLargest`, 200, getBody(validLength, "1234567", 0)},
		{`LengthTooLong`, 400, getBody(validLength, "12345678", 0)},

		{`SizeTooSmall`, 400, getBody(validSize, "", 2)},
		{`SizeSmallest`, 200, getBody(validSize, "", 3)},
		{`SizeLargest`, 200, getBody(validSize, "", 10)},
		{`SizeTooBig`, 400, getBody(validSize, "", 11)},

		{`ExclusiveTooSmall`, 400, getBody(exclusiveSize, "", 3)},
		{`ExclusiveSmallest`, 200, getBody(exclusiveSize, "", 4)},
		{`ExclusiveLargest`, 200, getBody(exclusiveSize, "", 9)},
		{`ExclusiveTooBig`, 400, getBody(exclusiveSize, "", 10)},

		{`ExclusiveOldTooSmall`, 400, getBody(exclusiveSizeOld, "", 3)},
		{`ExclusiveOldSmallest`, 200, getBody(exclusiveSizeOld, "", 4)},
		{`ExclusiveOldLargest`, 200, getBody(exclusiveSizeOld, "", 9)},
		{`ExclusiveOldTooBig`, 400, getBody(exclusiveSizeOld, "", 10)},

		{`NonExclusiveOldTooSmall`, 400, getBody(nonExclusiveSizeOld, "", 2)},
		{`NonExclusiveOldSmallest`, 200, getBody(nonExclusiveSizeOld, "", 3)},
		{`NonExclusiveOldLargest`, 200, getBody(nonExclusiveSizeOld, "", 10)},
		{`NonExclusiveOldTooBig`, 400, getBody(nonExclusiveSizeOld, "", 11)},

		{`patternSimpleSuccess`, 200, getBody(patternSimple, "aaa", 0)},
		{`patternSimpleFail`, 400, getBody(patternSimple, "a", 0)},
		{`patternWithNegativeLookaheadSuccess`, 200, getBody(patternWithNegativeLookahead, "1.1.1.1", 0)},
		{`patternWithNegativeLookaheadFail`, 400, getBody(patternWithNegativeLookahead, "1", 0)},

		{`enumStringSuccess`, 200, getBody(enumString, "Val1", 0)},
		{`enumStringSuccessWithSpace`, 200, getBody(enumString, "Val With Spaces", 0)},
		{`enumStringSuccessWithTabs`, 200, getBody(enumString, "Val\tWith\tTabs", 0)},
		{`enumStringFail`, 400, getBody(enumString, "val1", 0)},

		{`enumStringRefSuccess`, 200, getBody(enumStringRef, "Val1", 0)},
		{`enumStringRefSuccessWithSpace`, 200, getBody(enumStringRef, "Val With Spaces", 0)},
		{`enumStringRefSuccessWithTabs`, 200, getBody(enumStringRef, "Val\tWith\tTabs", 0)},
		{`enumStringRefFail`, 400, getBody(enumStringRef, "val1", 0)},

		{`enumIntSuccess`, 200, getBody(enumInt, "", 1)},
		{`enumIntFail`, 400, getBody(enumInt, "", 4)},

		{`LargeNumber`, 200, getBody(largeNumber, "", 9999999)},

		{`ArrayOfObjectDiveTooShort`, 400, getBody(arrayOfObject, "1", 0)},
		{`ArrayOfObjectDive`, 200, getBody(arrayOfObject, "12", 0)},

		{`ArrayOfEnumDiveInvalid`, 400, getBody(arrayOfEnum, "Invalid", 0)},
		{`ArrayOfEnumDive`, 200, getBody(arrayOfEnum, "Val1", 0)},

		{`FloatWithFractionalSmallest`, 200, getBodyFloat(floatWithFractional, 1)},
		{`FloatWithFractionalLargest`, 200, getBodyFloat(floatWithFractional, 999999999.99)},
		{`FloatWithFractionalTooLarge`, 400, getBodyFloat(floatWithFractional, 1000000000)},

		{`RefOfArrayValid`, 200, getBody(refOfArrayWithLength, "123", 0)},
		{`RefOfArrayTooLong`, 400, getBody(refOfArrayWithLength, "1234", 0)},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			gatewayTester.PostPingWithValidate().
				WithHeaders(getPopulatedHeader()).
				WithBody(test.body).
				ExpectResponseCode(test.code).
				Send()
		})
	}
}

func getHeader(vts varToSet, s string) map[string]string {
	ret := getPopulatedHeader()

	switch vts {
	case validLength:
		ret["headerLength"] = s
	case validLengthNil:
		delete(ret, "headerLength")
	case patternWithNegativeLookahead:
		ret["headerPattern"] = s
	}

	return ret
}

func TestValidate_HeaderParams(t *testing.T) {
	t.Parallel()
	gatewayTester := pingpongwithvalidate.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	for _, test := range []struct {
		name    string
		code    int
		headers map[string]string
	}{
		{`allPopulated`, 200, getPopulatedHeader()},

		{`missingRequired`, 400, getHeader(validLengthNil, "")},

		{`LengthTooShort`, 400, getHeader(validLength, "1")},
		{`LengthSmallest`, 200, getHeader(validLength, "12")},
		{`LengthLargest`, 200, getHeader(validLength, "1234567")},
		{`LengthTooLong`, 400, getHeader(validLength, "12345678")},

		{`patternWithNegativeLookaheadSuccess`, 200, getHeader(patternWithNegativeLookahead, "1.1.1.1")},
		{`patternWithNegativeLookaheadFail`, 400, getHeader(patternWithNegativeLookahead, "1")},
	} {
		test := test
		body := getPopulatedBody()
		t.Run(test.name, func(t *testing.T) {
			gatewayTester.PostPingWithValidate().
				WithHeaders(test.headers).
				WithBody(body).
				ExpectResponseCode(test.code).
				Send()
		})
	}
}

func TestValidate_PathParams(t *testing.T) {
	t.Parallel()
	gatewayTester := pingpongwithvalidate.NewTestServer(t, context.Background(), createService, "")
	defer gatewayTester.Close()

	for _, test := range []struct {
		name            string
		code            int
		pathWithPattern string
		pathWithLength  string
	}{
		{`allPopulated`, 200, "7d83d140bd56", "123"},

		{`LengthTooShort`, 400, "7d83d140bd56", "1"},
		{`LengthSmallest`, 200, "7d83d140bd56", "12"},
		{`LengthLargest`, 200, "7d83d140bd56", "1234567"},
		{`LengthTooLong`, 400, "7d83d140bd56", "12345678"},

		{`patternFail`, 400, "invalid", "123"},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			gatewayTester.GetPingPathParamWithValidate(test.pathWithPattern, test.pathWithLength).
				ExpectResponseCode(test.code).
				Send()
		})
	}
}
