package templatemap

import (
	"reflect"
	"strings"
	"text/template"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	gormLogger "gorm.io/gorm/logger"
)

var CoreFuncMap = template.FuncMap{
	"executeTemplate": ExecuteTemplate,
	"setValue":        SetValue,
	"getValue":        GetValue,
	"toSQL":           ToSQL,
	"exec":            Exec,
}

func getRepositoryFromVolume(volume VolumeInterface) RepositoryInterface {
	var r RepositoryInterface
	ok := volume.GetValue(REPOSITORY_KEY, &r)
	if !ok {
		err := errors.Errorf("not found repository  key %s in %#v", REPOSITORY_KEY, volume)
		panic(err)
	}
	return r
}

//ExecuteTemplate 模板中调用模板
func ExecuteTemplate(volume VolumeInterface, name string) (string, error) {
	var r = getRepositoryFromVolume(volume)
	out, err := r.ExecuteTemplate(name, volume)
	if err != nil {
		return "", err
	}
	out = strings.ReplaceAll(out, WINDOW_EOF, EOF)
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

func Exec(volume VolumeInterface, identifier string, s string) (string, error) {
	var provider ExecproviderInterface
	var r = getRepositoryFromVolume(volume)
	provider, ok := r.GetProvider(identifier)
	if !ok {
		err := errors.Errorf("not found provider by identifier: %s", identifier)
		return "", err
	}
	out, err := provider.Exec(identifier, s)
	if err != nil {
		return "", err
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
