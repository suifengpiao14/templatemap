package templatemap

import (
	"strings"

	"github.com/pkg/errors"
)

func RawSchema2jsonSchema(rawSchema string) *Schema {
	schema := new(Schema)
	rawSchema = strings.ReplaceAll(rawSchema, WINDOW_EOF, EOF)
	rawkArr := strings.Split(rawSchema, EOF)
	for _, raw := range rawkArr {
		raw = StandardizeSpaces(raw)
		pairArr := strings.Split(raw, ",")
		kvmap := make(map[string]interface{})
		fullname := ""
		for _, pair := range pairArr {
			kv := strings.SplitN(pair, ":", 2)
			if len(kv) != 2 {
				err := errors.Errorf("error pair format,except k:v ,got:%#v", pair)
				panic(err)
			}
			key := kv[0]
			value := kv[1]
			if key == "fullname" {
				fullname = value
			}
			kvmap[key] = value
		}
		schema.SetByFullName(fullname, kvmap)
	}
	return schema
}
