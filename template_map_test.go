package templatemap

import (
	"fmt"
	"testing"
)

func TestAddTemplateByDir(t *testing.T) {

	r := NewRepository(TemplatefuncMap)
	dir := "."
	err := r.AddTemplateByDir(dir)
	if err != nil {
		panic(err)
	}

	volume := Volume{
		"APIIDList": []int{1, 3, 4},
	}
	out, err := r.ExecuteTemplate("PaginateTotal", &volume)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	fmt.Println(volume)
}
