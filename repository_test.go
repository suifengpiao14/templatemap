package templatemap

import (
	"fmt"
	"testing"
)

func TestRepository(t *testing.T) {
	r := NewRepository()
	r.RegisterProvider("curl", &CURLExecProvider{})
	r.RegisterProvider("docapi_db", &DBExecProvider{LogLevel: SQL_LOG_LEVEL_DEBUG, DSN: "hjx:123456@tcp(106.53.100.222:3306)/docapi?charset=utf8&timeout=1s&readTimeout=5s&writeTimeout=5s&parseTime=False&loc=Local&multiStatements=true"})
	r.RegisterProvider("docapi_db2", &DBExecProvider{LogLevel: SQL_LOG_LEVEL_DEBUG, DSN: "hjx:123456@tcp(106.53.100.222:3306)/docapi?charset=utf8&timeout=1s&readTimeout=5s&writeTimeout=5s&parseTime=False&loc=Local&multiStatements=true"})

	err := r.AddTemplateByDir(".")
	if err != nil {
		panic(err)
	}
	volume := Volume{
		"PageIndex": "0",
		"PageSize":  "20",
	}
	_, err = r.ExecuteTemplate("getPaginate", &volume)
	if err != nil {
		panic(err)
	}

	jsonschema := `{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","properties":{"items":{"type":"array","items":{"type":"object","properties":{"sourceId":{"type":"string","src":"PaginateOut.#.api_id"},"sourceType":{"type":"string","src":"PaginateOut.#.name"},"url":{"type":"string","src":"PaginateOut.#.url"},"createTime":{"type":"string","src":"PaginateOut.#.created_at"},"updateTime":{"type":"string","src":"PaginateOut.#.updated_at"}},"required":["sourceId","sourceType","url","createTime","updateTime"]}},"pagination":{"type":"object","properties":{"total":{"type":"integer","src":"PaginateTotalOut"},"pageIndex":{"type":"string","src":"PageIndex"},"pageSize":{"type":"string","src":"PageSize"}},"required":["total","pageIndex","pageSize"]}},"required":["items","pagination"]}`

	paths, err := JsonSchema2Path(jsonschema)
	if err != nil {
		panic(err)
	}
	out, err := TransferData(&volume, paths)
	if err != nil {
		panic(err)
	}

	//fmt.Println(volume)
	fmt.Println(out)
}
