package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

const (
	DBName = "TimerNotification"
)

var dsn string

func BuildDSN() string {
	if dsn == "" {
		databaseConfig := &Config.DatabaseConfig
		dsn = fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8", databaseConfig.DBUser, databaseConfig.DBPassword, databaseConfig.DBHost, databaseConfig.DBPort, DBName)
	}
	return dsn
}

func GetDBConnection() (*sql.DB, error) {
	return sql.Open("mysql", BuildDSN())
}
