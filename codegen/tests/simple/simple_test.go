package simple

import (
	"bufio"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/convert"
	"github.com/anz-bank/sysl-go/restlib"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestHandler struct {
	reqH  http.Header
	respH http.Header
	s     int
}

func headerCpy(src http.Header) http.Header {
	dst := make(http.Header, len(src))

	for k, vv := range src {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		dst[k] = vv2
	}

	return dst
}

func (th *TestHandler) ValidGetStuffListHandlerStub(ctx context.Context, req *GetStuffListRequest, client GetStuffListClient) (*Stuff, error) {
	s := Stuff{
		InnerStuff: "response",
		ResponseStuff: Response{
			Data: ItemSet{
				M: map[string]Item{
					"John": {
						A1:   "CollinsSt",
						A2:   "LonasDaleSt",
						Name: "John",
					},
					"James": {
						A1:   "SpencerSt",
						A2:   "CollinsSt",
						Name: "James",
					},
				},
			},
		},
	}

	th.reqH = headerCpy(common.RequestHeaderFromContext(ctx))

	respH, status := common.RespHeaderAndStatusFromContext(ctx)
	th.respH = headerCpy(respH)
	th.s = status

	return &s, nil
}

func (th *TestHandler) ValidRawHandler(ctx context.Context, req *GetRawListRequest, client GetRawListClient) (*Str, error) {
	var s Str = "raw"

	th.reqH = headerCpy(common.RequestHeaderFromContext(ctx))

	respH, status := common.RespHeaderAndStatusFromContext(ctx)
	th.respH = headerCpy(respH)
	th.s = status

	return &s, nil
}

func (th *TestHandler) ValidRawIntHandler(ctx context.Context, req *GetRawIntListRequest, client GetRawIntListClient) (*Integer, error) {
	var s Integer = 123

	th.reqH = headerCpy(common.RequestHeaderFromContext(ctx))

	respH, status := common.RespHeaderAndStatusFromContext(ctx)
	th.respH = headerCpy(respH)
	th.s = status

	return &s, nil
}

func (th *TestHandler) InvalidHander(ctx context.Context, req *GetStuffListRequest, client GetStuffListClient) (*Stuff, error) {
	return nil, errors.New("invalid")
}

func callHandler(target string, si ServiceInterface) (*httptest.ResponseRecorder, *test.Hook) {
	cb := Callback{}

	sh := NewServiceHandler(cb, &si)

	r := httptest.NewRequest("GET", target, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Accept", "application/json")
	logger, hook := test.NewNullLogger()
	r = r.WithContext(common.LoggerToContext(context.Background(), logger, logrus.NewEntry(logger)))

	sh.GetStuffListHandler(w, r)

	return w, hook
}

func callRawHandler(target string, si ServiceInterface) (*httptest.ResponseRecorder, *test.Hook) {
	cb := Callback{}

	sh := NewServiceHandler(cb, &si)

	r := httptest.NewRequest("GET", target, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Accept", "application/json")
	logger, hook := test.NewNullLogger()
	r = r.WithContext(common.LoggerToContext(context.Background(), logger, logrus.NewEntry(logger)))

	sh.GetRawListHandler(w, r)

	return w, hook
}

func callRawIntHandler(target string, si ServiceInterface) (*httptest.ResponseRecorder, *test.Hook) {
	cb := Callback{}

	sh := NewServiceHandler(cb, &si)

	r := httptest.NewRequest("GET", target, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Accept", "application/json")
	logger, hook := test.NewNullLogger()
	r = r.WithContext(common.LoggerToContext(context.Background(), logger, logrus.NewEntry(logger)))

	sh.GetRawIntListHandler(w, r)

	return w, hook
}

func TestHandlerNotImplemented(t *testing.T) {
	w, hook := callHandler("http://example.com/stuff", ServiceInterface{})

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	require.JSONEq(t, `{"status":{"code":"9998", "description":"Internal Server Error"}}`, string(body))
	require.Equal(t, "ServerError(Kind=Internal Server Error, Message=not implemented, Cause=%!s(<nil>))", hook.LastEntry().Message)
}

func TestHandlerMissingEndpoint(t *testing.T) {
	w, hook := callHandler("http://example.com/gruff", ServiceInterface{})

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	require.JSONEq(t, `{"status":{"code":"9998", "description":"Internal Server Error"}}`, string(body))
	require.Equal(t, "ServerError(Kind=Internal Server Error, Message=not implemented, Cause=%!s(<nil>))", hook.LastEntry().Message)
}

func TestHandlerRequestHeaderInContext(t *testing.T) {
	th := TestHandler{}
	si := ServiceInterface{
		GetStuffList: th.ValidGetStuffListHandlerStub,
	}

	_, _ = callHandler("http://example.com/stuff", si)

	require.Equal(t, "application/json", th.reqH.Get("Accept"))
}

func TestHandlerResponseHeaderInContext(t *testing.T) {
	th := TestHandler{}
	si := ServiceInterface{
		GetStuffList: th.ValidGetStuffListHandlerStub,
	}

	_, _ = callHandler("http://example.com/stuff", si)

	require.Equal(t, 200, th.s)
	require.Equal(t, 0, len(th.respH))
}

func TestHandlerValid(t *testing.T) {
	th := TestHandler{}
	si := ServiceInterface{
		GetStuffList: th.ValidGetStuffListHandlerStub,
	}

	w, _ := callHandler("http://example.com/stuff", si)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	require.JSONEq(t, `{"emptyStuff":{}, "innerStuff":"response", "rawTimeStuff":"0001-01-01T00:00:00Z", "responseStuff":{"Data":{"M":{"James":{"A1":"SpencerSt", "A2":"CollinsSt"}, "John":{"A1":"CollinsSt", "A2":"LonasDaleSt"}}}}, "sensitiveStuff":"****************", "timeStuff":"0001-01-01T00:00:00.000+0000"}`, string(body))
}

func TestRawHandlerValid(t *testing.T) {
	th := TestHandler{}
	si := ServiceInterface{
		GetRawList: th.ValidRawHandler,
	}

	w, _ := callRawHandler("http://example.com/raw", si)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	require.JSONEq(t, `"raw"`, string(body))
}

func TestRawIntHandlerValid(t *testing.T) {
	th := TestHandler{}
	si := ServiceInterface{
		GetRawIntList: th.ValidRawIntHandler,
	}

	w, _ := callRawIntHandler("http://example.com/raw-int", si)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	require.JSONEq(t, `123`, string(body))
}

func TestHandlerDownstreamInvalid(t *testing.T) {
	t.Skip("Skipping due to missing dependency")
	th := TestHandler{}
	si := ServiceInterface{
		GetStuffList: th.InvalidHander,
	}

	w, hook := callHandler("http://example.com/stuff", si)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	require.JSONEq(t, `{"status":{"code":"1234", "description":"Unknown Error"}}`, string(body))
	require.Equal(t, "ServerError(Kind=Unexpected response from downstream services, Message=Downstream failure, Cause=invalid)", hook.LastEntry().Message)
}

func TestClientDecodesValidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test request parameters
		require.Equal(t, r.URL.String(), "/stuff")
		// Send response to be tested
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"innerStuff":"test"}`))
	}))
	client := server.Client()
	defer server.Close()

	c := Client{
		client: client,
		url:    server.URL,
	}

	logger, _ := test.NewNullLogger()
	ctx := common.LoggerToContext(context.Background(), logger, logrus.NewEntry(logger))

	result, err := c.GetStuffList(ctx, &GetStuffListRequest{})
	require.NoError(t, err)
	require.Equal(t, Stuff{InnerStuff: "test"}, *result)
}

func validQueryParamTest(t *testing.T, req GetStuffListRequest, query string) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test request parameters
		require.Equal(t, r.URL.String(), query)
		// Send response to be tested
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"innerStuff":"test"}`))
	}))
	client := server.Client()
	defer server.Close()

	c := Client{
		client: client,
		url:    server.URL,
	}

	logger, _ := test.NewNullLogger()
	ctx := common.LoggerToContext(context.Background(), logger, logrus.NewEntry(logger))

	_, err := c.GetStuffList(ctx, &req)
	require.NoError(t, err)
}

func validXMLMsgTest(t *testing.T, req PostStuffRequest, xmlBody string) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, len(xmlBody))
		w.Header().Add("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(200)
		_, err := r.Body.Read(body)
		require.Equal(t, err, io.EOF)
		_, _ = w.Write(body)
	}))
	client := server.Client()
	defer server.Close()

	c := Client{
		client: client,
		url:    server.URL,
	}

	logger, _ := test.NewNullLogger()
	reqHeader := http.Header{}
	reqHeader.Add("Content-Type", "text/xml; charset=utf-8")
	ctx := common.RequestHeaderToContext(common.LoggerToContext(context.Background(), logger,
		logrus.NewEntry(logger)), reqHeader)
	strRes, err := c.PostStuff(ctx, &req)
	assert.Equal(t, string(*strRes), xmlBody)
	require.NoError(t, err)
}

func TestClientPassesValidOptionalStringParam(t *testing.T) {
	testString := "test string"
	req := GetStuffListRequest{
		Dt: nil,
		St: &testString,
	}

	validQueryParamTest(t, req, "/stuff?st=test+string")
}

func TestClientPassesValidOptionalBoolParam(t *testing.T) {
	testBool := true
	req := GetStuffListRequest{
		Dt: nil,
		St: nil,
		Bt: &testBool,
	}

	validQueryParamTest(t, req, "/stuff?bt=true")
}

func TestClientPassesValidOptionalIntParam(t *testing.T) {
	testInt := int64(42)

	req := GetStuffListRequest{
		Dt: nil,
		St: nil,
		Bt: nil,
		It: &testInt,
	}

	validQueryParamTest(t, req, "/stuff?it=42")
}

func TestClientPassesValidOptionalDatetimeParam(t *testing.T) {
	req := GetStuffListRequest{
		Dt: &convert.JSONTime{Time: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)},
		St: nil,
		Bt: nil,
		It: nil,
	}

	validQueryParamTest(t, req, "/stuff?dt=2009-11-10+23%3A00%3A00+%2B0000+UTC")
}

func TestClientPassesAllValidOptionalParams(t *testing.T) {
	testInt := int64(42)
	testBool := true
	testString := "test string"

	req := GetStuffListRequest{
		Dt: &convert.JSONTime{Time: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)},
		St: &testString,
		Bt: &testBool,
		It: &testInt,
	}

	validQueryParamTest(t, req, "/stuff?bt=true&dt=2009-11-10+23%3A00%3A00+%2B0000+UTC&it=42&st=test+string")
}

func TestClient_PassesXMLBody(t *testing.T) {
	xmlBody := "<test><stuff>test stuff</stuff></test>"
	req := PostStuffRequest{
		Request: Str(xmlBody),
	}
	validXMLMsgTest(t, req, xmlBody)
}

func TestSensitive(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.Error(Stuff{InnerStuff: "innerStuff", SensitiveStuff: common.NewSensitiveString("sensitiveStuff")})
	require.Equal(t, "{{} innerStuff 0001-01-01 00:00:00 +0000 UTC {{map[]}} **************** 0001-01-01 00:00:00 +0000 UTC}", hook.LastEntry().Message)
}

func TestTimeFormat(t *testing.T) {
	stuff := Stuff{}
	isStdTime := func(s interface{}) bool {
		switch s.(type) {
		case time.Time:
			return true
		default:
			return false
		}
	}

	isConvertTime := func(s interface{}) bool {
		switch s.(type) {
		case convert.JSONTime:
			return true
		default:
			return false
		}
	}

	require.True(t, isStdTime(stuff.RawTimeStuff))
	require.False(t, isConvertTime(stuff.RawTimeStuff))

	require.True(t, isConvertTime(stuff.TimeStuff))
	require.False(t, isStdTime(stuff.TimeStuff))
}

func TestCommentsPassed(t *testing.T) {
	fh, err := os.Open("./types.go")
	require.NoError(t, err)
	defer fh.Close()

	s := bufio.NewScanner(fh)

	foundComment := false

	for s.Scan() {
		if strings.Contains(s.Text(), "// Stuff just some stuff") {
			foundComment = true
			break
		}
	}

	require.True(t, foundComment)
}

func bodylessClientServer(statusToReturn int) (*Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send response to be tested
		w.Header().Add("Context", `{"jsonField":"jsonVal"}`)
		w.WriteHeader(statusToReturn)
	}))
	client := server.Client()

	return &Client{
		client: client,
		url:    server.URL,
	}, server
}

func TestJustOKReturnsHeaders(t *testing.T) {
	c, s := bodylessClientServer(200)
	defer s.Close()

	logger, _ := test.NewNullLogger()
	ctx := common.LoggerToContext(context.Background(), logger, logrus.NewEntry(logger))

	h, err := c.GetJustReturnOkList(ctx, &GetJustReturnOkListRequest{})
	require.NoError(t, err)
	require.Equal(t, `{"jsonField":"jsonVal"}`, h.Get("Context"))
}

func TestJustErrorPutsHeadersInError(t *testing.T) {
	c, s := bodylessClientServer(400)
	defer s.Close()

	logger, _ := test.NewNullLogger()
	ctx := common.LoggerToContext(context.Background(), logger, logrus.NewEntry(logger))

	err := c.GetJustReturnErrorList(ctx, &GetJustReturnErrorListRequest{})
	require.Error(t, err)
	resp := err.(*common.ServerError).Cause.(*restlib.HTTPResult)
	require.Equal(t, `{"jsonField":"jsonVal"}`, resp.HTTPResponse.Header.Get("Context"))
}

func TestJustOKAndJustErrorReturnsHeadersWhenOK(t *testing.T) {
	c, s := bodylessClientServer(200)
	defer s.Close()

	logger, _ := test.NewNullLogger()
	ctx := common.LoggerToContext(context.Background(), logger, logrus.NewEntry(logger))

	h, err := c.GetJustOkAndJustErrorList(ctx, &GetJustOkAndJustErrorListRequest{})
	require.NoError(t, err)
	require.Equal(t, `{"jsonField":"jsonVal"}`, h.Get("Context"))
}

func TestJustOKAndJustErrorPutsHeadersInErrorWhenError(t *testing.T) {
	c, s := bodylessClientServer(400)
	defer s.Close()

	logger, _ := test.NewNullLogger()
	ctx := common.LoggerToContext(context.Background(), logger, logrus.NewEntry(logger))

	h, err := c.GetJustOkAndJustErrorList(ctx, &GetJustOkAndJustErrorListRequest{})
	require.Error(t, err)
	require.Nil(t, h)
	resp := err.(*common.ServerError).Cause.(*restlib.HTTPResult)
	require.Equal(t, `{"jsonField":"jsonVal"}`, resp.HTTPResponse.Header.Get("Context"))
}

func TestOKTypeAndJustErrorPutsHeadersInErrorWhenError(t *testing.T) {
	c, s := bodylessClientServer(400)
	defer s.Close()

	logger, _ := test.NewNullLogger()
	ctx := common.LoggerToContext(context.Background(), logger, logrus.NewEntry(logger))

	h, err := c.GetOkTypeAndJustErrorList(ctx, &GetOkTypeAndJustErrorListRequest{})
	require.Error(t, err)
	require.Nil(t, h)
	resp := err.(*common.ServerError).Cause.(*restlib.HTTPResult)
	require.Equal(t, `{"jsonField":"jsonVal"}`, resp.HTTPResponse.Header.Get("Context"))
}
