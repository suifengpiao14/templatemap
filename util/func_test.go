package util

import (
	"fmt"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

func TestColumn2Row(t *testing.T) {
	jsonStr := `
	{"name":["张三","李四","王五"],"age":[1,2,3]}
	`
	jsonStr = `
	{"accountId":["34845c07-e1f9-4c1f-864c-30543ea7eb2e","42087ce6-0492-41ae-b996-b5e351bab3df","12243"],"createdAt":["2022-10-15 16:11:32","2022-10-15 15:31:06","2022-10-15 11:36:52"],"deletedAt":["","",""],"name":["admin2","admin2","admin1"],"password":["123456","123456","123456"],"phone":["15999646794","15999646794","15999646793"],"role":["admin","admin","admin"],"updatedAt":["2022-10-15 16:11:32","2022-10-15 15:31:06","2022-10-15 11:36:52"],"userId":["3","2","1"]}
	`

	out := Column2Row(jsonStr)
	fmt.Println(out)
}

func TestRow2Column(t *testing.T) {

	jsonStr := `
	[{"name":"张三","age":1},{"name":"李四","age":2},{"name":"王五","age":3}]
	`

	out := Row2Column(jsonStr)
	fmt.Println(out)
}

func TestValid(t *testing.T) {
	input := `{"id":"1","price":"10"}`
	schema := `{"$id":"input","required":["config"],"properties":{"config":{"properties":{"desc":{"type":"string","dst":"Fdesc"},"id":{"type":"string","format":"number","dst":"Fid"},"name":{"type":"string","dst":"Fname"},"os":{"type":"string","format":"number","dst":"Fos"},"price":{"type":"string","format":"number","dst":"Fprice"},"projectId":{"type":"string","format":"number","dst":"FprojectID"},"rewardKey":{"type":"string","dst":"FrewardKey"},"sortValue":{"type":"string","format":"number","dst":"FsortValue"},"valid":{"type":"string","format":"number","dst":"Fvalid"}},"type":"object","required":["desc","id","name","os","price","projectId","rewardKey","sortValue","valid"]}},"type":"object"}`
	schemaLoader := gojsonschema.NewStringLoader(schema)
	err := Validate(input, schemaLoader)
	if err != nil {
		panic(err)
	}
}
