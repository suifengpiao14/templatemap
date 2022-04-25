package templatemap

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var DriverName = "mysql"

var DB_SOURCE = ""

const (
	SQL_TYPE_SELECT = "SELECT"
	SQL_TYPE_OTHER  = "OTHER"
)

type DBExecProvider struct {
	DSN    string
	db     *sqlx.DB
	dbOnce sync.Once
}

func (p *DBExecProvider) Exec(identifier string, s string) (string, error) {
	return dbProvider(p, s)
}

// GetDb is a signal DB
func (p *DBExecProvider) GetDb() *sqlx.DB {
	if p.db == nil {

		if p.DSN == "" {
			err := errors.Errorf("DBExecProvider %#v DNS is null ", p)
			panic(err)
		}
		p.dbOnce.Do(func() {
			db, err := sqlx.Open(DriverName, p.DSN)
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
	fmt.Println(sqls)
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
	rows, err := db.Queryx(sqls)
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
			err := rows.MapScan(record)
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
