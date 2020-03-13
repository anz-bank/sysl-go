package dbendpoints

import (
	"context"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/anz-bank/sysl-go-comms/common"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func callHandler(target string, si ServiceInterface) (*httptest.ResponseRecorder, *test.Hook) {
	cb := Callback{}

	sh := NewServiceHandler(cb, &si)

	r := httptest.NewRequest("GET", target, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	logger, hook := test.NewNullLogger()
	r = r.WithContext(common.LoggerToContext(context.Background(), logger, logrus.NewEntry(logger)))

	sh.GetCompanyLocationListHandler(w, r)

	return w, hook
}

func TestHandler_Valid(t *testing.T) {
	siImpl := NewDefaultDbEndpointsImpl()
	si := ServiceInterface{
		GetCompanyLocationList: siImpl.GetCompanyLocationList(),
	}

	w, _ := callHandler("http://example.com/company/location?companyName=ANZ&deptLoc=MELB", si)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	expectedResponse := `{"companies":[{"abnNumber":1,"companyCountry":"AU","companyName":"ANZ","departments":[{"deptId":1,"deptLoc":"MELB","deptName":"FINANCE"},{"deptId":4,"deptLoc":"MELB","deptName":"POSTAL"},{"deptId":6,"deptLoc":"MELB","deptName":"IT"},{"deptId":7,"deptLoc":"MELB","deptName":"POSTAL"}]}],"message":"OK"}`
	require.JSONEq(t, expectedResponse, string(body))
}
