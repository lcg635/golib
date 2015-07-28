package storage

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/doug-martin/goqu.v2"
	_ "gopkg.in/doug-martin/goqu.v2/adapters/mysql"
)

type MysqlClient struct {
	Db *goqu.Database
}

func NewMysqlClient(dataSourceName string) *MysqlClient {
	mysqlDb, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	return &MysqlClient{
		Db: goqu.New("mysql", mysqlDb),
	}
}
