package templatemap

import (
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"goa.design/goa/v3/codegen"
	gormLogger "gorm.io/gorm/logger"
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
	"toSQL":           ToSQL,
	"execSQL":         ExecSQL,
}

const IN_INDEX = "__inIndex"
const DB_PROVIDER_KEY = "__dbProvidor" // 数据库执行器

type DBProvidorInterface interface {
	Exec(DBIdentifier string, sql string) (interface{}, error)
}

type DBProviderFunc func(DBIdentifier string, sql string) (interface{}, error)

func (f DBProviderFunc) Exec(DBIdentifier string, sql string) (interface{}, error) {
	// 调用f函数本体
	return f(DBIdentifier, sql)
}

func Add() {

}

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

func ExecSQL(volume Volume, DBIdentifier string, sql string) (interface{}, error) {
	var dbProvidor DBProvidorInterface
	ok := volume.GetValue(DB_PROVIDER_KEY, &dbProvidor)
	if !ok {
		err := errors.Errorf("not found db providor by identifier: %s", DBIdentifier)
		return nil, err
	}
	out, err := dbProvidor.Exec(DBIdentifier, sql)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func ToSQL(volume VolumeInterface, namedSQL string) (string, error) {
	data, err := getNamedData(volume)
	if err != nil {
		return "", err
	}
	statment, arguments, err := sqlx.Named(namedSQL, data)
	if err != nil {
		err = errors.WithStack(err)
		return "", err
	}
	sql := gormLogger.ExplainSQL(statment, nil, `'`, arguments...)
	return sql, nil
}

func getNamedData(data interface{}) (out map[string]interface{}, err error) {
	out = make(map[string]interface{})
	if data == nil {
		return
	}
	dataI, ok := data.(*interface{})
	if ok {
		data = *dataI
	}
	mapOut, ok := data.(map[string]interface{})
	if ok {
		out = mapOut
		return
	}
	mapOutRef, ok := data.(*map[string]interface{})
	if ok {
		out = *mapOutRef
		return
	}
	if mapOut, ok := data.(Volume); ok {
		out = mapOut
		return
	}
	if mapOutRef, ok := data.(*Volume); ok {
		out = *mapOutRef
		return
	}

	v := reflect.Indirect(reflect.ValueOf(data))

	if v.Kind() != reflect.Struct {
		return
	}
	vt := v.Type()
	// 提取结构体field字段
	fieldNum := v.NumField()
	for i := 0; i < fieldNum; i++ {
		fv := v.Field(i)
		ft := fv.Type()
		fname := vt.Field(i).Name
		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
			ft = fv.Type()
		}
		ftk := ft.Kind()
		switch ftk {
		case reflect.Int:
			out[fname] = fv.Int()
		case reflect.Int64:
			out[fname] = int64(fv.Int())
		case reflect.Float64:
			out[fname] = fv.Float()
		case reflect.String:
			out[fname] = fv.String()
		case reflect.Struct, reflect.Map:
			subOut, err := getNamedData(fv.Interface())
			if err != nil {
				return out, err
			}
			for k, v := range subOut {
				if _, ok := out[k]; !ok {
					out[k] = v
				}
			}

		default:
			out[fname] = fv.Interface()
		}
	}
	return
}
