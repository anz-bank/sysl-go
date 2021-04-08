package restlib

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/common"
	"github.com/stretchr/testify/require"
)

type OkType struct {
	Test string `json:"test"`
}

type ErrorType struct {
	Test2 string `json:"test2"`
}

type BytesType []byte

const (
	okJSON    = `{ "test":"test string" }`
	errorJSON = `{ "test2":"test string 2" }`
)

func TestUnmarshalPanicOnNilResponse(t *testing.T) {
	require.Panics(t, func() { _, _ = unmarshal(nil, nil, nil) })
}

func TestUnmarshalBadJson(t *testing.T) {
	result, err := unmarshal(&http.Response{}, []byte(`{ "bad-JSON`), &OkType{})
	require.Error(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTPResponse)
	require.NotNil(t, result.Body)
	require.Nil(t, result.Response)
}

func TestUnmarshalNilBodyOK(t *testing.T) {
	result, err := unmarshal(&http.Response{}, nil, OkType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTPResponse)
	require.Nil(t, result.Body)
	require.Nil(t, result.Response)
}

func TestUnmarshalPointerBodyOK(t *testing.T) {
	result, err := unmarshal(&http.Response{}, nil, &OkType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTPResponse)
	require.Nil(t, result.Body)
	require.NotNil(t, result.Response)
}

func TestUnmarshalEmptyBodyOK(t *testing.T) {
	result, err := unmarshal(&http.Response{}, make([]byte, 0), OkType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTPResponse)
	require.IsType(t, []byte{}, result.Body)
	require.Nil(t, result.Response)
}

func TestUnmarshalBytesContent(t *testing.T) {
	header := map[string][]string{
		"Content-Type": {"image/png"},
	}
	var image = []byte{1, 2}
	result, err := unmarshal(&http.Response{Header: header}, image, &BytesType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTPResponse)
	require.NotNil(t, result.Body)
	_, ok := result.Response.(*BytesType)
	require.True(t, ok)
}

func TestUnmarshalRawStringContent(t *testing.T) {
	header := map[string][]string{
		"Content-Type": {"text/plain"},
	}
	var response = ""
	result, err := unmarshal(&http.Response{Header: header}, []byte("hello"), &response)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTPResponse)
	require.NotNil(t, result.Body)
	_, ok := result.Response.(*string)
	require.True(t, ok)
}

func TestUnmarshalRawBytesContent(t *testing.T) {
	header := map[string][]string{
		"Content-Type": {"application/octet-stream"},
	}
	var image = []byte{1, 2}
	var response []byte = nil
	result, err := unmarshal(&http.Response{Header: header}, image, &response)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTPResponse)
	require.NotNil(t, result.Body)
	_, ok := result.Response.(*[]byte)
	require.True(t, ok)
}

func TestUnmarshalNilTypeOK(t *testing.T) {
	result, err := unmarshal(&http.Response{}, []byte(okJSON), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.HTTPResponse)
	require.NotNil(t, result.Body)
	require.Nil(t, result.Response)
}

func TestUnmarshalWrongJSONOK(t *testing.T) {
	result, err := unmarshal(&http.Response{}, []byte(errorJSON), &OkType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.IsType(t, &OkType{}, result.Response)
}

func TestUnmarshalAliasString(t *testing.T) {
	type Str string
	var OkStrType Str
	result, err := unmarshal(&http.Response{}, []byte(errorJSON), &OkStrType)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.IsType(t, &OkStrType, result.Response)
}

func TestDoHTTPRequestOkType(t *testing.T) {
	srv := common.NewHTTPTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(okJSON))
	}))
	defer srv.Close()

	result, err := DoHTTPRequest(context.Background(), srv.Client(), "GET", srv.URL, nil, make([]string, 0), &OkType{}, &ErrorType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.IsType(t, &OkType{}, result.Response)
}

func TestDoHTTPRequest204Response(t *testing.T) {
	srv := common.NewHTTPTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
		_, _ = w.Write(nil)
	}))
	defer srv.Close()

	result, err := DoHTTPRequest(context.Background(), srv.Client(), "GET", srv.URL, nil, make([]string, 0), &OkType{}, &ErrorType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.IsType(t, &OkType{}, result.Response)
}

func TestDoHTTPRequest204ResponseGZIP(t *testing.T) {
	srv := common.NewHTTPTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Encoding", "gzip")
		w.WriteHeader(204)
		_, _ = w.Write(nil)
	}))
	defer srv.Close()

	result, err := DoHTTPRequest(context.Background(), srv.Client(), "GET", srv.URL, nil, make([]string, 0), &OkType{}, &ErrorType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.IsType(t, &OkType{}, result.Response)
}

func TestDoHTTPRequestErrorType(t *testing.T) {
	srv := common.NewHTTPTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestDoHTTPRequestRightTypeWrongJSON(t *testing.T) {
	srv := common.NewHTTPTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(errorJSON))
	}))
	defer srv.Close()

	result, err := DoHTTPRequest(context.Background(), srv.Client(), "GET", srv.URL, nil, make([]string, 0), &OkType{}, &ErrorType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.IsType(t, &OkType{}, result.Response)
}

func TestDoHTTPRequestXMLBody(t *testing.T) {
	xmlBody := `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ond="http://www.qas.com/OnDemand-2011-03"><soapenv:Header><ond:QAQueryHeader></ond:QAQueryHeader></soapenv:Header><soapenv:Body><ond:QASearch><ond:Country>AUS</ond:Country><ond:Engine>Intuitive</ond:Engine><!--Optional:--><ond:Layout>QADefault</ond:Layout><ond:Search>5 lyg</ond:Search><!--Optional:--><ond:FormattedAddressInPicklist>false</ond:FormattedAddressInPicklist></ond:QASearch></soapenv:Body></soapenv:Envelope>`
	srv := common.NewHTTPTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestDoHTTPRequestSendStructAsUrlEncodedBody(t *testing.T) {
	type BananaRequest struct {
		Banana     string
		BananaType string
		ExpiresAt  time.Time
	}

	req := &BananaRequest{
		Banana:     "ripe",
		BananaType: "wrapped",
		ExpiresAt:  time.Date(2021, time.February, 10, 0, 0, 0, 0, time.UTC),
	}

	expectedURLEncodedData := []byte(`Banana=ripe&BananaType=wrapped&ExpiresAt=2021-02-10T00%3A00%3A00Z`)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(errorJSON))
			return
		}
		if !bytes.Equal(expectedURLEncodedData, data) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(errorJSON))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(okJSON))
	}))
	defer srv.Close()

	reqHeader := http.Header{}
	reqHeader.Add("Content-Type", "application/x-www-form-urlencoded")
	ctx := common.RequestHeaderToContext(context.Background(), reqHeader)

	result, err := DoHTTPRequest(ctx, srv.Client(), "POST", srv.URL, req, make([]string, 0), &OkType{}, &ErrorType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	responseObj, ok := result.Response.(*OkType)
	require.True(t, ok)
	expectedResponseObj := &OkType{Test: "test string"}
	require.Equal(t, expectedResponseObj, responseObj)
}

func TestDoHTTPRequestSendStructAsUrlEncodedBodyWithCharset(t *testing.T) {
	type BananaRequest struct {
		Banana     string
		BananaType string
		ExpiresAt  time.Time
	}

	req := &BananaRequest{
		Banana:     "ripe",
		BananaType: "wrapped",
		ExpiresAt:  time.Date(2021, time.February, 10, 0, 0, 0, 0, time.UTC),
	}

	expectedURLEncodedData := []byte(`Banana=ripe&BananaType=wrapped&ExpiresAt=2021-02-10T00%3A00%3A00Z`)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(errorJSON))
			return
		}
		if !bytes.Equal(expectedURLEncodedData, data) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(errorJSON))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(okJSON))
	}))
	defer srv.Close()

	reqHeader := http.Header{}
	reqHeader.Add("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	ctx := common.RequestHeaderToContext(context.Background(), reqHeader)

	result, err := DoHTTPRequest(ctx, srv.Client(), "POST", srv.URL, req, make([]string, 0), &OkType{}, &ErrorType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	responseObj, ok := result.Response.(*OkType)
	require.True(t, ok)
	expectedResponseObj := &OkType{Test: "test string"}
	require.Equal(t, expectedResponseObj, responseObj)
}

func TestDoHTTPRequestSendStructWithCustomUrlFieldTagsAsUrlEncodedBody(t *testing.T) {
	type BananaRequest struct {
		Banana                  string    `url:"banana"`
		BananaType              string    `url:"banana_type"`
		ExpiresAt               time.Time `url:"expires_at"`
		ProprietaryBananaSecret string    `url:"-"`
	}

	// Ref: https://pkg.go.dev/github.com/google/go-querystring/query
	req := &BananaRequest{
		Banana:                  "ripe",
		BananaType:              "wrapped",
		ExpiresAt:               time.Date(2021, time.February, 10, 0, 0, 0, 0, time.UTC),
		ProprietaryBananaSecret: "THIS MUST NOT BE SENT OVER THE WIRE",
	}

	expectedURLEncodedData := []byte(`banana=ripe&banana_type=wrapped&expires_at=2021-02-10T00%3A00%3A00Z`)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(errorJSON))
			return
		}
		if !bytes.Equal(expectedURLEncodedData, data) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(errorJSON))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(okJSON))
	}))
	defer srv.Close()

	reqHeader := http.Header{}
	reqHeader.Add("Content-Type", "application/x-www-form-urlencoded")
	ctx := common.RequestHeaderToContext(context.Background(), reqHeader)

	result, err := DoHTTPRequest(ctx, srv.Client(), "POST", srv.URL, req, make([]string, 0), &OkType{}, &ErrorType{})
	require.NoError(t, err)
	require.NotNil(t, result)
	responseObj, ok := result.Response.(*OkType)
	require.True(t, ok)
	expectedResponseObj := &OkType{Test: "test string"}
	require.Equal(t, expectedResponseObj, responseObj)
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

func TestSendHTTPResponseContentTypeImage(t *testing.T) {
	// Given
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "image/jpeg")

	// When
	data := &BytesType{1, 2}
	SendHTTPResponse(recorder, 200, data)

	// Then
	result := recorder.Result()
	require.NotNil(t, result)
	require.Equal(t, 200, result.StatusCode)
	b, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2}, b)
}

func TestSendHTTPResponseContentTypeTextPlain(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "text/plain")

	data := "Plain text"
	SendHTTPResponse(recorder, 200, data)

	result := recorder.Result()
	require.NotNil(t, result)
	require.Equal(t, 200, result.StatusCode)
	b, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	require.NoError(t, err)
	require.Equal(t, []byte(data), b)
}

func TestSendHTTPResponseContentTypeTextHtml(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "text/html")

	data := "Plain text"
	SendHTTPResponse(recorder, 200, data)

	result := recorder.Result()
	require.NotNil(t, result)
	require.Equal(t, 200, result.StatusCode)
	b, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	require.NoError(t, err)
	require.Equal(t, []byte(data), b)
}

func TestSendHTTPResponseContentTypeOctetStream(t *testing.T) {
	recorder := httptest.NewRecorder()
	recorder.Header().Set("Content-Type", "application/octet-stream")

	data := []byte("Encoded")
	SendHTTPResponse(recorder, 200, data)

	result := recorder.Result()
	require.NotNil(t, result)
	require.Equal(t, 200, result.StatusCode)
	b, err := ioutil.ReadAll(result.Body)
	defer result.Body.Close()
	require.NoError(t, err)
	require.Equal(t, data, b)
}

func TestRestResultContextWithoutProvision(t *testing.T) {
	srv := common.NewHTTPTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(okJSON))
	}))
	defer srv.Close()

	ctx := context.Background()
	result, err := DoHTTPRequest(ctx, srv.Client(), "GET", srv.URL, nil, make([]string, 0), &OkType{}, &ErrorType{})
	OnRestResultHTTPResult(ctx, result, err)
	restResult := common.GetRestResult(ctx)
	require.Nil(t, restResult)
}

func TestRestResultContextOnSuccess(t *testing.T) {
	srv := common.NewHTTPTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(okJSON))
	}))
	defer srv.Close()

	ctx := context.Background()
	ctx = common.ProvisionRestResult(ctx)
	result, err := DoHTTPRequest(ctx, srv.Client(), "GET", srv.URL, nil, make([]string, 0), &OkType{}, &ErrorType{})
	OnRestResultHTTPResult(ctx, result, err)
	restResult := common.GetRestResult(ctx)
	require.NotNil(t, restResult)

	require.NotNil(t, restResult)
	require.Equal(t, 200, restResult.StatusCode)
	require.NotEmpty(t, restResult.Headers)
	require.Equal(t, []byte(okJSON), restResult.Body)
}

func TestRestResultContextOnError(t *testing.T) {
	srv := common.NewHTTPTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(errorJSON))
	}))
	defer srv.Close()

	ctx := context.Background()
	ctx = common.ProvisionRestResult(ctx)
	result, err := DoHTTPRequest(ctx, srv.Client(), "GET", srv.URL, nil, make([]string, 0), &OkType{}, &ErrorType{})
	OnRestResultHTTPResult(ctx, result, err)
	restResult := common.GetRestResult(ctx)
	require.NotNil(t, restResult)

	require.NotNil(t, restResult)
	require.Equal(t, 400, restResult.StatusCode)
	require.NotEmpty(t, restResult.Headers)
	require.Equal(t, []byte(errorJSON), restResult.Body)
}

func TestMarshalRequestBodyTextPlain(t *testing.T) {
	content := "Hello world"
	reader, err := marshalRequestBody("text/plain", content)
	require.Nil(t, err)
	require.NotNil(t, reader)
	marshalled, err := ioutil.ReadAll(reader)
	require.Nil(t, err)
	require.Equal(t, content, string(marshalled))
}

func TestMarshalRequestBodyTextHtml(t *testing.T) {
	content := "Hello world"
	reader, err := marshalRequestBody("text/html", content)
	require.Nil(t, err)
	require.NotNil(t, reader)
	marshalled, err := ioutil.ReadAll(reader)
	require.Nil(t, err)
	require.Equal(t, content, string(marshalled))
}

func TestMarshalRequestBodyOctetStream(t *testing.T) {
	content := []byte("Hello world")
	reader, err := marshalRequestBody("application/octet-stream", content)
	require.Nil(t, err)
	require.NotNil(t, reader)
	marshalled, err := ioutil.ReadAll(reader)
	require.Nil(t, err)
	require.Equal(t, content, marshalled)
}
