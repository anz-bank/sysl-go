package dbendpoints

import (
	"context"
	"fmt"
)

func (s *DefaultDbEndpointsImpl) GetCompanyLocationList() func(ctx context.Context, req *GetCompanyLocationListRequest, client GetCompanyLocationListClient) (*GetCompanyLocationResponse, error) {
	return func(ctx context.Context, req *GetCompanyLocationListRequest, client GetCompanyLocationListClient) (*GetCompanyLocationResponse, error) {
		queryStmt := client.retrievebycompanyandlocation
		rows, err := queryStmt.QueryContext(ctx, req.DeptLoc, req.CompanyName)
		if err != nil {
			fmt.Println("################ error Received - " + err.Error())
			return nil, err
		}
		companiesMap := map[int64]Company{}
		departmentsMap := map[int64][]Department{}
		var departments []Department
		var prevAbnNum, curCompanyIndex int64
		for rows.Next() {
			if rows.Err() != nil {
				fmt.Println("################ error Received - " + err.Error())
				return nil, err
			}
			var abnnumber, deptid int64
			var companyname, companycountry, deptname, deptloc string
			_ = rows.Scan(&abnnumber, &companyname, &companycountry, &deptid, &deptname, &deptloc)
			department := Department{
				DeptId:   deptid,
				DeptName: deptname,
				DeptLoc:  deptloc,
			}
			if abnnumber != prevAbnNum {
				departments = []Department{}
				departments = append(departments, department)
				company := Company{
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
		companiesList := []Company{}
		for abn := range companiesMap {
			departmentsList := departmentsMap[abn]
			comp := Company{
				AbnNumber:      companiesMap[abn].AbnNumber,
				CompanyName:    companiesMap[abn].CompanyName,
				CompanyCountry: companiesMap[abn].CompanyCountry,
				Departments:    departmentsList,
			}
			companiesList = append(companiesList, comp)
		}

		getCompanyLocationResponse := GetCompanyLocationResponse{
			Companies: companiesList,
			Message:   "OK",
		}
		return &getCompanyLocationResponse, nil
	}
}
