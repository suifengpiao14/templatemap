package templatemap

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	gormLogger "gorm.io/gorm/logger"
)

var CoreFuncMap = template.FuncMap{
	"executeTemplate": ExecuteTemplate,
	"setValue":        SetValue,
	"getValue":        GetValue,
	"toSQL":           ToSQL,
	"exec":            Exec,
	"execSQLTpl":      ExecSQLTpl,
	"execCURLTpl":     ExecCURLTpl,
	"gjsonGet":        gjson.Get,
	"sjsonSet":        sjson.Set,
	"sjsonSetRaw":     sjson.SetRaw,
	"transfer":        Transfer,
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
func ExecuteTemplate(volume VolumeInterface, name string) string {
	var r = getRepositoryFromVolume(volume)
	out, err := r.ExecuteTemplate(name, volume)
	if err != nil {
		panic(err) // ExecuteTemplate 函数可能嵌套很多层，抛出错误值后有可能被当成正常值处理，所以此处直接panic 退出，保留原始错误输出
	}
	out = strings.ReplaceAll(out, WINDOW_EOF, EOF)
	return out
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

func Exec(volume VolumeInterface, tplName string, s string) string {
	var provider ExecproviderInterface
	var r = getRepositoryFromVolume(volume)
	provider, ok := r.GetProvider(tplName)
	if !ok {
		err := errors.Errorf("not found provider by template name : %s", tplName)
		panic(err)
	}
	out, err := provider.Exec(tplName, s)
	if err != nil {
		panic(err)
	}
	return out
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

func ExecCURLTpl(volume VolumeInterface, templateName string) error {
	tplOut := ExecuteTemplate(volume, templateName)
	out := Exec(volume, templateName, tplOut)
	storeKey := fmt.Sprintf("%sOut", templateName)
	volume.SetValue(storeKey, out)
	return nil
}

func ExecSQLTpl(volume VolumeInterface, templateName string) string {
	//{{executeTemplate . "Paginate"|toSQL . | exec . "docapi_db2"|setValue . }}
	tplOut := ExecuteTemplate(volume, templateName)
	tplOut = StandardizeSpaces(tplOut)
	if tplOut == "" {
		err := errors.Errorf("template :%s output empty", tplOut)
		panic(err)
	}
	sql, err := ToSQL(volume, tplOut)
	if err != nil {
		panic(err)
	}
	sqlKey := fmt.Sprintf("%sSQL", templateName)
	volume.SetValue(sqlKey, sql)
	out := Exec(volume, templateName, sql)
	if err != nil {
		panic(err)
	}
	storeKey := fmt.Sprintf("%sOut", templateName)
	volume.SetValue(storeKey, out)
	return "" // 符合模板函数，至少一个输出结构
}

func Transfer(volume Volume, dstSchema string) (interface{}, error) {
	return nil, nil
}

func JsonSchema2Path(jsonschema string) (TransferPaths, error) {
	var schema Schema
	err := json.Unmarshal([]byte(jsonschema), &schema)
	if err != nil {
		return nil, err
	}
	schema.Init()
	out := schema.GetTransferPaths()

	return out, nil

}

func TransferData(volume VolumeInterface, transferPaths TransferPaths) (string, error) {
	out := ""
	for _, tp := range transferPaths {
		var v interface{}
		var err error
		ok := volume.GetValue(tp.Src, &v)
		if !ok {
			err := errors.Errorf("not found %s data from volume %#v", tp.Src, volume)
			return "", err
		}
		if strings.Contains(tp.Dst, "#") {
			arr, ok := v.([]interface{})
			if !ok {
				err = errors.Errorf("")
				return "", err
			}
			// 初始化数组元素(可以确保空数组的写入)
			/* 			keyArr := strings.SplitN(tp.Dst, "#", 2)
			   			out, err = sjson.Set(out, keyArr[0], "[]")
			   			if err != nil {
			   				err = errors.WithStack(err)
			   				return "", err
			   			} */
			for index, val := range arr {
				path := strings.ReplaceAll(tp.Dst, "#", strconv.Itoa(index))
				out, err = sjson.Set(out, path, val)
				if err != nil {
					err = errors.WithStack(err)
					return "", err
				}
			}
			continue

		}
		out, err = sjson.Set(out, tp.Dst, v)
		if err != nil {
			err = errors.WithStack(err)
			return "", err
		}
	}
	return out, nil
}
