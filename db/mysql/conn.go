package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)
var DBConn *sql.DB

func init() {
	DBConn,_ = sql.Open("mysql","root:)Yonglun1003@tcp(127.0.0.1:3306)/file_store?charset=utf8")
	DBConn.SetMaxOpenConns(1000)
	err := DBConn.Ping()
	if err != nil {
		fmt.Println("Failed to connect to mysql, err:",err.Error())
		os.Exit(1)
	}
}

// DBCONN返回数据库连接对象
func DBCONN() *sql.DB{
	return DBConn
}

func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		//将行数据保存到record字典
		err := rows.Scan(scanArgs...)
		checkErr(err)

		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		records = append(records, record)
	}
	return records
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
