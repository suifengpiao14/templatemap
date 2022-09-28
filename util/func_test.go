package util

import (
	"fmt"
	"testing"
)

func TestColumn2Row(t *testing.T) {
	jsonStr := `
	{"name":["张三","李四","王五"],"age":[1,2,3]}
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
