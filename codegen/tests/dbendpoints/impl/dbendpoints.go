package impl

import (
	"context"
	"fmt"
	db "github.com/anz-bank/sysl-go/codegen/arrai/tests/dbendpoints"
)

func GetCompanyLocationList(ctx context.Context, req *db.GetCompanyLocationListRequest, client db.GetCompanyLocationListClient) (*db.GetCompanyLocationResponse, error) {
	queryStmt := client.Retrievebycompanyandlocation
	rows, err := queryStmt.QueryContext(ctx, req.DeptLoc, req.CompanyName)
	if err != nil {
		fmt.Println("################ error Received - " + err.Error())
		return nil, err
	}
	companiesMap := map[int64]db.Company{}
	departmentsMap := map[int64][]db.Department{}
	var departments []db.Department
	var prevAbnNum, curCompanyIndex int64
	for rows.Next() {
		if rows.Err() != nil {
			fmt.Println("################ error Received - " + err.Error())
			return nil, err
		}
		var abnnumber, deptid int64
		var companyname, companycountry, deptname, deptloc string
		_ = rows.Scan(&abnnumber, &companyname, &companycountry, &deptid, &deptname, &deptloc)
		department := db.Department{
			DeptId:   deptid,
			DeptName: deptname,
			DeptLoc:  deptloc,
		}
		if abnnumber != prevAbnNum {
			departments = []db.Department{}
			departments = append(departments, department)
			company := db.Company{
				AbnNumber:      abnnumber,
				CompanyName:    companyname,
				CompanyCountry: companycountry,
			}
			companiesMap[abnnumber] = company
			departmentsMap[abnnumber] = departments
			curCompanyIndex++
			prevAbnNum = abnnumber
		} else {
			departments = append(departments, department)
			departmentsMap[abnnumber] = departments
		}
	}
	companiesList := []db.Company{}
	for abn := range companiesMap {
		departmentsList := departmentsMap[abn]
		comp := db.Company{
			AbnNumber:      companiesMap[abn].AbnNumber,
			CompanyName:    companiesMap[abn].CompanyName,
			CompanyCountry: companiesMap[abn].CompanyCountry,
			Departments:    departmentsList,
		}
		companiesList = append(companiesList, comp)
	}

	getCompanyLocationResponse := db.GetCompanyLocationResponse{
		Companies: companiesList,
		Message:   "OK",
	}
	return &getCompanyLocationResponse, nil
}
