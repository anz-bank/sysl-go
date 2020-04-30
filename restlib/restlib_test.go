package restlib

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anz-bank/sysl-go/common"
	"github.com/stretchr/testify/require"
)

type OkType struct {
	Test string `json:"test"`
}

type ErrorType struct {
	Test2 string `json:"test2"`
}

const (
	okJSON    = `{ "test":"test string" }`
	errorJSON = `{ "test2":"test string 2" }`
)

func Test_unmarshalPanicOnNilResponse(t *testing.T) {
	require.Panics(t, func() { _, _ = unmarshal(nil, nil, nil) })
}

func Test_unmarshalNilBodyOK(t *testing.T) {
	result, err := unmarshal(&http.Response{}, nil, OkType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTPResponse)
	require.Nil(t, result.Body)
	require.Nil(t, result.Response)
}

func Test_unmarshalPointerBodyOK(t *testing.T) {
	result, err := unmarshal(&http.Response{}, nil, &OkType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTPResponse)
	require.Nil(t, result.Body)
	require.NotNil(t, result.Response)
}

func Test_unmarshalEmptyBodyOK(t *testing.T) {
	result, err := unmarshal(&http.Response{}, make([]byte, 0), OkType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTPResponse)
	require.IsType(t, []byte{}, result.Body)
	require.Nil(t, result.Response)
}

func Test_unmarshalNilTypeOK(t *testing.T) {
	result, err := unmarshal(&http.Response{}, []byte(okJSON), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTPResponse)
	require.NotNil(t, result.Body)
	require.Nil(t, result.Response)
}

func Test_unmarshalWrongJSONOK(t *testing.T) {
	result, err := unmarshal(&http.Response{}, []byte(errorJSON), &OkType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.IsType(t, &OkType{}, result.Response)
}

func Test_unmarshalAliasString(t *testing.T) {
	type Str string
	var OkStrType Str
	result, err := unmarshal(&http.Response{}, []byte(errorJSON), &OkStrType)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.IsType(t, &OkStrType, result.Response)
}

func Test_DoHTTPRequestOkType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(okJSON))
	}))
	defer srv.Close()

	result, err := DoHTTPRequest(context.Background(), srv.Client(), "GET", srv.URL, nil, make([]string, 0), &OkType{}, &ErrorType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.IsType(t, &OkType{}, result.Response)
}

func Test_DoHTTPRequest204Response(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
		_, _ = w.Write(nil)
	}))
	defer srv.Close()

	result, err := DoHTTPRequest(context.Background(), srv.Client(), "GET", srv.URL, nil, make([]string, 0), &OkType{}, &ErrorType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.IsType(t, &OkType{}, result.Response)
}

func Test_DoHTTPRequestErrorType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(errorJSON))
	}))
	defer srv.Close()

	result, err := DoHTTPRequest(context.Background(), srv.Client(), "GET", srv.URL, nil, make([]string, 0), &OkType{}, &ErrorType{})
	require.Error(t, err)
	require.IsType(t, &HTTPResult{}, err)
	require.Nil(t, result)
	require.IsType(t, &ErrorType{}, err.(*HTTPResult).Response)
}

func Test_DoHTTPRequestRightTypeWrongJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(errorJSON))
	}))
	defer srv.Close()

	result, err := DoHTTPRequest(context.Background(), srv.Client(), "GET", srv.URL, nil, make([]string, 0), &OkType{}, &ErrorType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.IsType(t, &OkType{}, result.Response)
}

func Test_DoHTTPRequestXMLBody(t *testing.T) {
	xmlBody := `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ond="http://www.qas.com/OnDemand-2011-03"><soapenv:Header><ond:QAQueryHeader></ond:QAQueryHeader></soapenv:Header><soapenv:Body><ond:QASearch><ond:Country>AUS</ond:Country><ond:Engine>Intuitive</ond:Engine><!--Optional:--><ond:Layout>QADefault</ond:Layout><ond:Search>5 lyg</ond:Search><!--Optional:--><ond:FormattedAddressInPicklist>false</ond:FormattedAddressInPicklist></ond:QASearch></soapenv:Body></soapenv:Envelope>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, len(xmlBody))
		w.Header().Add("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(200)
		_, err := r.Body.Read(body)
		require.Equal(t, err, io.EOF)
		_, _ = w.Write(body)
	}))
	defer srv.Close()
	reqHeader := http.Header{}
	reqHeader.Add("Content-Type", "text/xml; charset=utf-8")
	ctx := common.RequestHeaderToContext(context.Background(), reqHeader)
	result, err := DoHTTPRequest(ctx, srv.Client(), "POST", srv.URL, xmlBody, make([]string, 0), &OkType{}, &ErrorType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	strRes, isString := result.Response.(string)
	require.True(t, isString)
	require.True(t, xmlBody == strRes)
}

type testResp struct {
	Data string `json:"jdata" xml:"xdata"`
}

func TestSendHTTPResponseJSONBody(t *testing.T) {
	// Given
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "application/json")

	resp := testResp{Data: "test"}

	// When
	SendHTTPResponse(recorder, 200, resp)

	// Then
	result := recorder.Result()
	require.NotNil(t, result)
	require.Equal(t, 200, result.StatusCode)
	b, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	require.NoError(t, err)
	require.Equal(t, "{\"jdata\":\"test\"}\n", string(b))
}

func TestSendHTTPResponseXMLBody(t *testing.T) {
	// Given
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "text/xml; charset=utf-8")

	resp := testResp{Data: "test"}

	// When
	SendHTTPResponse(recorder, 200, resp)

	// Then
	result := recorder.Result()
	require.NotNil(t, result)
	require.Equal(t, 200, result.StatusCode)
	b, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	require.NoError(t, err)
	require.Equal(t, "<testResp><xdata>test</xdata></testResp>", string(b))
}

func TestSendHTTPResponseBinaryBody(t *testing.T) {
	// Given
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "application/octet-stream")

	// When
	data := []byte("test binary data")
	SendHTTPResponse(recorder, 200, data)

	// Then
	result := recorder.Result()
	require.NotNil(t, result)
	require.Equal(t, 200, result.StatusCode)
	b, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	require.NoError(t, err)
	require.Equal(t, data, b)
}

type ByteWrapper []byte

func TestSendHTTPResponseBinaryBody2(t *testing.T) {
	// Given
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "application/pdf")

	// When
	data := ByteWrapper("test binary data")
	SendHTTPResponse(recorder, 200, (*[]byte)(&data))

	// Then
	result := recorder.Result()
	require.NotNil(t, result)
	require.Equal(t, 200, result.StatusCode)
	b, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	require.NoError(t, err)
	require.Equal(t, ([]byte)(data), b)
}
