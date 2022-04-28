package templatemap

import (
	"fmt"
	"testing"
)

func TestRepository(t *testing.T) {
	r := NewRepository()
	tplnames := r.AddTemplateByDir(".")
	provider := &DBExecProvider{Config: DBExecProviderConfig{LogLevel: LOG_LEVEL_DEBUG, DSN: "hjx:123456@tcp(106.53.100.222:3306)/docapi?charset=utf8&timeout=1s&readTimeout=5s&writeTimeout=5s&parseTime=False&loc=Local&multiStatements=true"}}
	r.RegisterProvider(provider, tplnames...)
	volume := Volume{
		"PageIndex": "0",
		"PageSize":  "20",
	}
	execOut, err := r.ExecuteTemplate("getPaginate", &volume)
	if err != nil {
		panic(err)
	}
	fmt.Println(execOut)

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

func TestConvertType(t *testing.T) {
	src := 10
	var dst bool
	ok := convertType(&dst, src)
	fmt.Println(ok)
	fmt.Println(dst)
}
