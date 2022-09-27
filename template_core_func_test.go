package templatemap

import (
	"fmt"
	"testing"
)

func TestJsonSchema2Path(t *testing.T) {
	jsonschema := `{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","properties":{"items":{"type":"array","items":{"type":"object","properties":{"sourceId":{"type":"string","src":"PaginateOut.#.api_id"},"sourceType":{"type":"string","src":"PaginateOut.#.name"},"url":{"type":"string","src":"PaginateOut.#.url"},"createTime":{"type":"string","src":"PaginateOut.#.created_at"},"updateTime":{"type":"string","src":"PaginateOut.#.updated_at"}},"required":["sourceId","sourceType","url","createTime","updateTime"]}},"pagination":{"type":"object","properties":{"total":{"type":"integer","src":"PaginateTotalOut"},"pageIndex":{"type":"string","src":"PageIndex"},"pageSize":{"type":"string","src":"PageSize"}},"required":["total","pageIndex","pageSize"]}},"required":["items","pagination"]}`

	schema := NewJsonSchema(jsonschema)

	for _, path := range schema.GetTransferPaths() {
		fmt.Println(*path)
	}

}

func TestTransferData(t *testing.T) {
	out := ""
	v := "0"
	err := Add2json(&out, "pageIndex", "integer", v)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}

func TestReversalJsonArr(t *testing.T) {

	jsonStr := `
	{"name":["张三","李四","王五"],"age":[1,2,3]}
	`
	volume := volumeMap{}
	key := "object.items"
	volume.SetValue(key, jsonStr)
	ReversalJsonArr(&volume, key)
	var out string
	volume.GetValue(key, &out)
	fmt.Println(out)
}
