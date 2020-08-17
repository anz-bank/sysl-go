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
		`DROP TABLE IF EXISTS Account;`,
		`CREATE TABLE Account (accountID INTEGER PRIMARY KEY, balance NUMBER);`,
	}
	db, err := sql.Open("sqlite3", "default.db")
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
