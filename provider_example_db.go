package templatemap

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

var DriverName = "mysql"

const (
	SQL_TYPE_SELECT     = "SELECT"
	SQL_TYPE_OTHER      = "OTHER"
	SQL_LOG_LEVEL_DEBUG = "debug"
	SQL_LOG_LEVEL_INFO  = "info"
	SQL_LOG_LEVEL_ERROR = "error"
)

type DBExecProvider struct {
	DSN      string
	logLevel string
	db       *sql.DB
	dbOnce   sync.Once
}

func (p *DBExecProvider) Exec(identifier string, s string) (string, error) {
	return dbProvider(p, s)
}

// GetDb is a signal DB
func (p *DBExecProvider) GetDb() *sql.DB {
	if p.db == nil {

		if p.DSN == "" {
			err := errors.Errorf("DBExecProvider %#v DNS is null ", p)
			panic(err)
		}
		p.dbOnce.Do(func() {
			db, err := sql.Open(DriverName, p.DSN)
			if err != nil {
				panic(err)
			}
			p.db = db

		})
	}
	return p.db
}

//SQLType 判断 sql  属于那种类型
func SQLType(sqls string) string {
	sqlArr := strings.Split(sqls, EOF)
	selectLen := len(SQL_TYPE_SELECT)
	for _, sql := range sqlArr {
		if len(sql) < selectLen {
			continue
		}
		pre := sql[:selectLen]
		if strings.ToUpper(pre) == SQL_TYPE_SELECT {
			return SQL_TYPE_SELECT
		}
	}
	return SQL_TYPE_OTHER
}

func dbProvider(p *DBExecProvider, sqls string) (string, error) {
	sqls = StandardizeSpaces(TrimSpaces(sqls)) // 格式化sql语句
	sqlType := SQLType(sqls)
	db := p.GetDb()
	if p.logLevel == SQL_LOG_LEVEL_DEBUG {
		fmt.Println(sqls)
	}
	if sqlType != SQL_TYPE_SELECT {
		res, err := db.Exec(sqls)
		if err != nil {
			return "", err
		}
		lastInsertId, _ := res.LastInsertId()
		if lastInsertId > 0 {
			return strconv.FormatInt(lastInsertId, 10), nil
		}
		rowsAffected, _ := res.RowsAffected()
		return strconv.FormatInt(rowsAffected, 10), nil
	}
	rows, err := db.Query(sqls)
	if err != nil {
		return "", err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			panic(err)
		}
	}()
	allResult := make([][]map[string]string, 0)
	for {
		records := make([]map[string]string, 0)
		for rows.Next() {
			var record = make(map[string]interface{})
			var recordStr = make(map[string]string)
			err := MapScan(*rows, record)
			if err != nil {
				return "", err
			}
			for k, v := range record {
				if v == nil {
					recordStr[k] = ""
				} else {
					recordStr[k] = fmt.Sprintf("%s", v)
				}
			}
			records = append(records, recordStr)
		}
		allResult = append(allResult, records)
		if !rows.NextResultSet() {
			break
		}
	}

	var jsonByte []byte
	if len(allResult) == 1 {
		result := allResult[0]
		if len(result) == 1 && len(result[0]) == 1 {
			row := result[0]
			for _, val := range row {
				return val, nil // 只有一个值时，直接返回值本身
			}
		}
		jsonByte, err = json.Marshal(allResult[0])
		if err != nil {
			return "", err
		}
	} else {
		jsonByte, err = json.Marshal(allResult)
		if err != nil {
			return "", err
		}
	}
	out := string(jsonByte)
	return out, nil
}

//MapScan copy sqlx
func MapScan(r sql.Rows, dest map[string]interface{}) error {
	// ignore r.started, since we needn't use reflect for anything.
	columns, err := r.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	err = r.Scan(values...)
	if err != nil {
		return err
	}

	for i, column := range columns {
		dest[column] = *(values[i].(*interface{}))
	}

	return r.Err()
}
