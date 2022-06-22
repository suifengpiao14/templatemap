package templatemap

import (
	"fmt"
	"testing"
)

func TestRepository(t *testing.T) {
	r := NewRepository()
	tplnames := r.AddTemplateByDir(".")
	provider := &DBExecProvider{Config: DBExecProviderConfig{LogLevel: LOG_LEVEL_DEBUG, DSN: "hjx:123456@tcp(106.53.100.222:3306)/docapi?charset=utf8&timeout=1s&readTimeout=5s&writeTimeout=5s&parseTime=False&loc=Local&multiStatements=true"}}
	for _, tplName := range tplnames {
		meta := TemplateMeta{
			Name:         tplName,
			ExecProvider: provider,
		}
		r.RegisterMeta(tplName, &meta)
	}
	volume := volumeMap{
		"PageIndex": "0",
		"PageSize":  "20",
	}
	tplName := "getPaginate"
	out, err := r.ExecuteTemplate(tplName, &volume)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}

func TestConvertType(t *testing.T) {
	src := 10
	var dst bool
	ok := convertType(&dst, src)
	fmt.Println(ok)
	fmt.Println(dst)
}
