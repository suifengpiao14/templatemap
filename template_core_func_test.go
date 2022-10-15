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

func TestListPadIndex(t *testing.T) {
	out := ListPadIndex(10)
	for i := range out {
		fmt.Println(i)
	}
}

func TestGetSetColumn2Row(t *testing.T) {

	jsonStr := `
	{"name":["张三","李四","王五"],"age":[1,2,3]}
	`
	volume := volumeMap{}
	key := "object.items"
	volume.SetValue(key, jsonStr)
	GetSetColumn2Row(&volume, key)
	var out string
	volume.GetValue(key, &out)
	fmt.Println(out)
}

func TestGetSetColumn2Row2(t *testing.T) {
	jsonStr := `
	{"items":{"accountId":["34845c07-e1f9-4c1f-864c-30543ea7eb2e","42087ce6-0492-41ae-b996-b5e351bab3df","12243"],"createdAt":["2022-10-15 16:11:32","2022-10-15 15:31:06","2022-10-15 11:36:52"],"deletedAt":["","",""],"name":["admin2","admin2","admin1"],"password":["123456","123456","123456"],"phone":["15999646794","15999646794","15999646793"],"role":["admin","admin","admin"],"updatedAt":["2022-10-15 16:11:32","2022-10-15 15:31:06","2022-10-15 11:36:52"],"userId":["3","2","1"]},"total":"3"}
	`
	volume := volumeMap{}
	key := "output"
	volume.SetValue(key, jsonStr)
	GetSetColumn2Row(&volume, "output.items")
	var out string
	volume.GetValue(key, &out)
	fmt.Println(out)
}

func TestGetSetColumn2Row3(t *testing.T) {
	jsonStr := `
	{"items":{"accountId":["34845c07-e1f9-4c1f-864c-30543ea7eb2e","42087ce6-0492-41ae-b996-b5e351bab3df","12243"],"createdAt":["2022-10-15 16:11:32","2022-10-15 15:31:06","2022-10-15 11:36:52"],"deletedAt":["","",""],"name":["admin2","admin2","admin1"],"password":["123456","123456","123456"],"phone":["15999646794","15999646794","15999646793"],"role":["admin","admin","admin"],"updatedAt":["2022-10-15 16:11:32","2022-10-15 15:31:06","2022-10-15 11:36:52"],"userId":["3","2","1"]},"total":"3"}
	`
	volume := volumeMap{}
	key := "output"
	key2 := "output.items1"
	volume.SetValue(key, jsonStr)
	GetSetColumn2Row(&volume, key2)
	var out string
	volume.GetValue(key2, &out)
	fmt.Println(out)
}

func TestGetSetRow2Column(t *testing.T) {

	jsonStr := `
	[{"name":"张三","age":1},{"name":"李四","age":2},{"name":"王五","age":3}]
	`
	volume := volumeMap{}
	key := "object.items"
	volume.SetValue(key, jsonStr)
	GetSetRow2Column(&volume, key)
	var out string
	volume.GetValue(key, &out)
	fmt.Println(out)
}
