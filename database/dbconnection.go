package database

import (
	"database/sql"
)

/*func GetDBHandle() (*sql.DB, error) {
	dbInfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", "sysl", "sysl", "sysl_db")
	db, err := sql.Open("postgres", dbInfo)
	if err != nil {
		fmt.Println("Error Received - " + err.Error())
		return nil, err
	}
	return db, nil
}*/

func GetDBHandle() (*sql.DB, error) {
	batch := []string{
		`CREATE TABLE company (abnnumber VARCHAR(50) PRIMARY KEY,companyname VARCHAR(50) ,companycountry VARCHAR(50));`,
		`CREATE TABLE department (deptid INTEGER PRIMARY KEY,deptname VARCHAR(50) ,deptloc VARCHAR(50) ,abn VARCHAR(50));`,
		`insert into company(abnnumber,companyname,companycountry) values('1','ANZ','AU');`,
		`insert into company(abnnumber,companyname,companycountry) values('2','NAB','AU');`,
		`insert into company(abnnumber,companyname,companycountry) values('3','CBA','AU');`,
		`insert into company(abnnumber,companyname,companycountry) values('4','STG','NZ');`,
		`insert into company(abnnumber,companyname,companycountry) values('5','BOM','NZ');`,
		`insert into department(deptid,deptname,deptloc,abn) values(1,'FINANCE','MELB',1);`,
		`insert into department(deptid,deptname,deptloc,abn) values(2,'HR','SYD',1);`,
		`insert into department(deptid,deptname,deptloc,abn) values(3,'IT','SYD',1);`,
		`insert into department(deptid,deptname,deptloc,abn) values(4,'POSTAL','MELB',1);`,
		`insert into department(deptid,deptname,deptloc,abn) values(5,'FINANCE','SYD',1);`,
		`insert into department(deptid,deptname,deptloc,abn) values(6,'IT','MELB',1);`,
		`insert into department(deptid,deptname,deptloc,abn) values(7,'POSTAL','MELB',1);`,
		`insert into department(deptid,deptname,deptloc,abn) values(8,'TRAVEL','SYD',1);`,
	}
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	for _, b := range batch {
		_, err = db.Exec(b)
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}
