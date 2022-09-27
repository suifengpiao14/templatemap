package util

import (
	"fmt"
	"testing"
)

func TestReversal(t *testing.T) {
	jsonStr := `
	{"name":["张三","李四","王五"],"age":[1,2,3]}
	`
	out := Reversal(jsonStr)
	fmt.Println(out)
}
