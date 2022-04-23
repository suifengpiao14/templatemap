package templatemap

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"goa.design/goa/v3/codegen"
)

var TemplatefuncMap = template.FuncMap{
	"zeroTime":        ZeroTime,
	"currentTime":     CurrentTime,
	"permanentTime":   PermanentTime,
	"contains":        strings.Contains,
	"newPreComma":     NewPreComma,
	"in":              In,
	"toCamel":         ToCamel,
	"toLowerCamel":    ToLowerCamel,
	"snakeCase":       SnakeCase,
	"executeTemplate": ExecuteTemplate,
	"setValue":        SetValue,
	"getValue":        GetValue,
}

const IN_INDEX = "__inIndex"

func ZeroTime(volume VolumeInterface) (string, error) {
	named := "ZeroTime"
	placeholder := ":" + named
	value := "0000-00-00 00:00:00"
	volume.SetValue(named, value)
	return placeholder, nil
}

func CurrentTime(volume VolumeInterface) (string, error) {
	named := "CurrentTime"
	placeholder := ":" + named
	value := time.Now().Format("2006-01-02 15:04:05")
	volume.SetValue(named, value)
	return placeholder, nil
}

func PermanentTime(volume VolumeInterface) (string, error) {
	named := "PermanentTime"
	placeholder := ":" + named
	value := "3000-12-31 23:59:59"
	volume.SetValue(named, value)
	return placeholder, nil
}

type preComma struct {
	comma string
}

func NewPreComma() *preComma {
	return &preComma{}
}

func (c *preComma) PreComma() string {
	out := c.comma
	c.comma = ","
	return out
}

func In(volume VolumeInterface, data interface{}) (str string, err error) {
	placeholders := make([]string, 0)
	inIndexKey := IN_INDEX
	var inIndex int
	ok := volume.GetValue(inIndexKey, &inIndex)
	if !ok {
		inIndex = 0
	}

	v := reflect.Indirect(reflect.ValueOf(data))

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		num := v.Len()
		for i := 0; i < num; i++ {
			inIndex++
			named := fmt.Sprintf("in_%d", inIndex)
			placeholder := ":" + named
			placeholders = append(placeholders, placeholder)
			volume.SetValue(named, v.Index(i).Interface())
		}

	case reflect.String:
		arr := strings.Split(v.String(), ",")
		num := len(arr)
		for i := 0; i < num; i++ {
			inIndex++
			named := fmt.Sprintf("in_%d", inIndex)
			placeholder := ":" + named
			placeholders = append(placeholders, placeholder)
			volume.SetValue(named, arr[i])
		}
	default:
		err = fmt.Errorf("want slice/array/string ,have %s", v.Kind().String())
		if err != nil {
			return "", err
		}
	}
	volume.SetValue(inIndexKey, inIndex) // 更新InIndex_
	str = strings.Join(placeholders, ",")
	return str, nil

}

// 封装 goa.design/goa/v3/codegen 方便后续可定制
func ToCamel(name string) string {
	return codegen.CamelCase(name, true, true)
}

func ToLowerCamel(name string) string {
	return codegen.CamelCase(name, false, true)
}

func SnakeCase(name string) string {
	return codegen.SnakeCase(name)
}

//ExecuteTemplate 模板中调用模板
func ExecuteTemplate(volume VolumeInterface, name string) (string, error) {
	var r repository
	ok := volume.GetValue(REPOSITORY_KEY, &r)
	if !ok {
		err := errors.Errorf("not found key %s in %#v", REPOSITORY_KEY, name)
		return "", err
	}
	out, err := r.ExecuteTemplate(name, volume)
	if err != nil {
		return "", err
	}
	return out, nil
}

func SetValue(volume VolumeInterface, key string, value interface{}) string { // SetValue 返回空字符，不对模板产生新输出
	volume.SetValue(key, value)
	return ""
}

func GetValue(volume VolumeInterface, key string) interface{} {
	var value interface{}
	volume.GetValue(key, &value)
	return value
}
