package templatemap

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRepository(t *testing.T) {
	r := NewRepository()
	r.RegisterProvider("curl", ExecProviderFuncCurl)
	r.RegisterProvider("docapi_db", &DBExecProvider{DSN: "hjx:123456@tcp(106.53.100.222:3306)/docapi?charset=utf8&timeout=1s&readTimeout=5s&writeTimeout=5s&parseTime=False&loc=Local&multiStatements=true"})
	r.RegisterProvider("docapi_db2", &DBExecProvider{DSN: "hjx:123456@tcp(106.53.100.222:3306)/docapi?charset=utf8&timeout=1s&readTimeout=5s&writeTimeout=5s&parseTime=False&loc=Local&multiStatements=true"})

	err := r.AddTemplateByDir(".")
	if err != nil {
		panic(err)
	}
	volume := Volume{
		"PageIndex": "0",
		"PageSize":  "20",
	}
	out, err := r.ExecuteTemplate("getPaginate", &volume)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	fmt.Println(volume)

}

func TestAddTemplateByDir(t *testing.T) {

	r := NewRepository()
	dir := "."
	err := r.AddTemplateByDir(dir)
	if err != nil {
		panic(err)
	}

	volume := Volume{
		"APIIDList":       []int{1, 3, 4},
		EXEC_PROVIDER_KEY: ExecProviderFunc(DBProvider1),
		"PageIndex":       "3",
		"PageSize":        "20",
	}
	out, err := r.ExecuteTemplate("PaginateTotal", &volume)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	fmt.Println(volume)
}

func DBProvider1(dbID string, sql string) (string, error) {
	out := []map[string]interface{}{
		{
			"ID":   1,
			"Name": "张三",
			"Age":  23,
		},
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
