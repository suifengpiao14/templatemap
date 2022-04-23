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
		"APIIDList":     []int{1, 3, 4},
		DB_PROVIDER_KEY: DBProviderFunc(DBProvider),
		"PageIndex":     "3",
		"PageSize":      "20",
	}
	out, err := r.ExecuteTemplate("PaginateTotal", &volume)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	fmt.Println(volume)
}

func DBProvider(dbID string, sql string) (interface{}, error) {
	out := []map[string]interface{}{
		{
			"ID":   1,
			"Name": "张三",
			"Age":  23,
		},
	}
	return out, nil
}
