package dbendpoints_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/codegen/arrai/tests/dbendpoints"
	"github.com/anz-bank/sysl-go/codegen/tests/dbendpoints/impl"
	"github.com/anz-bank/sysl-go/common"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestHandler_Valid(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	sh := dbendpoints.NewServiceHandler(
		common.Callback{},
		&dbendpoints.ServiceInterface{GetCompanyLocationList: impl.GetCompanyLocationList},
		db,
	)

	rows := sqlmock.NewRows(
		[]string{"abnnumber", "companyname", "companycountry", "deptid", "deptname", "deptloc"}).
		AddRow(1, "ANZ", "AU", 1, "FINANCE", "MELB").
		AddRow(1, "ANZ", "AU", 4, "POSTAL", "MELB").
		AddRow(1, "ANZ", "AU", 6, "IT", "MELB").
		AddRow(1, "ANZ", "AU", 7, "POSTAL", "MELB")
	mock.ExpectBegin()
	mock.ExpectPrepare("select company.abnnumber, company.companyname, company.companycountry, department.deptid, department.deptname, department.deptloc from company JOIN department ON company.abnnumber=department.abn WHERE department.deptloc=\\? and company.companyname=\\? order by company.abnnumber;").
		ExpectQuery().
		WithArgs("MELB", "ANZ").
		WillReturnRows(rows)
	mock.ExpectCommit()

	resBody := callHandler(sh, "http://example.com/company/location?companyName=ANZ&deptLoc=MELB")

	expectedResponse := `{"companies":[{"abnNumber":1,"companyCountry":"AU","companyName":"ANZ","departments":[{"deptId":1,"deptLoc":"MELB","deptName":"FINANCE"},{"deptId":4,"deptLoc":"MELB","deptName":"POSTAL"},{"deptId":6,"deptLoc":"MELB","deptName":"IT"},{"deptId":7,"deptLoc":"MELB","deptName":"POSTAL"}]}],"message":"OK"}`
	require.JSONEq(t, expectedResponse, resBody)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func callHandler(sh *dbendpoints.ServiceHandler, target string) string {
	r := httptest.NewRequest("GET", target, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	logger, _ := test.NewNullLogger()
	ctx, _ := context.WithTimeout(context.Background(), 300*time.Second)
	r = r.WithContext(common.LoggerToContext(ctx, logger, logrus.NewEntry(logger)))

	sh.GetCompanyLocationListHandler(w, r)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}
