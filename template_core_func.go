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
	var r = getRepositoryFromVolume(volume)
	meta, ok := r.GetMeta(tplName)
	if !ok {
		err := errors.Errorf("not found meta  by template name : %s", tplName)
		panic(err)
	}

	provider := meta.ExecProvider
	if provider == nil {
		err := errors.Errorf("meta:%v provider must be set", meta)
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
	if mapOut, ok := data.(volumeMap); ok {
		out = mapOut
		return
	}
	if mapOutRef, ok := data.(*volumeMap); ok {
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

func Transfer(volume volumeMap, dstSchema string) (interface{}, error) {
	return nil, nil
}

//TransferFiledToVolume 根据TransferPaths 提炼数据到volume 根节点下，主要用于不同接口间输入输出数据的承接
func TransferFiledToVolume(volume VolumeInterface, p TransferPaths) {
	data, err := TransferDataFromVolume(volume, p)
	if err != nil {
		panic(err)
	}
	var input map[string]interface{}
	err = json.Unmarshal([]byte(data), &input)
	if err != nil {
		panic(err)
	}
	for k, v := range input {
		volume.SetValue(k, v)
	}
}

func TransferDataFromVolume(volume VolumeInterface, transferPaths TransferPaths) (string, error) {
	out := ""
	for _, tp := range transferPaths {
		var v interface{}
		var err error
		ok := volume.GetValue(tp.Src, &v)
		if !ok {
			optionalTp := tp.GetOptionalTransferPath()
			if optionalTp == nil {
				err := errors.Errorf("not found %s data from volume %#v", tp.Src, volume)
				return "", err
			}
			continue
		}
		err = Add2json(&out, tp.Dst, tp.DstType, v)
		if err != nil {
			return "", err
		}
	}
	return out, nil
}

// 从json字符串中提取部分值，形成新的json字符串
func TransferJson(input string, transferPaths TransferPaths) (string, error) {
	out := ""
	for _, tp := range transferPaths {
		var v interface{}
		var err error
		result := gjson.Get(input, tp.Src)
		if !result.Exists() {
			v = tp.Default
		} else {
			v = result.String()
		}
		err = Add2json(&out, tp.Dst, tp.DstType, v)
		if err != nil {
			return "", err
		}
	}
	return out, nil
}

//Add2json 数据转换(将go数据写入到json字符串中)
func Add2json(s *string, dstPath string, dstType string, v interface{}) error {
	var err error
	if strings.Contains(dstPath, "#") {
		arr, ok := v.([]interface{})
		if !ok { // todo 此处只考虑了，从json字符串中提取数据，在设置到新的json字符串方式，对于
			err = errors.Errorf("")
			return err
		}
		if len(arr) == 0 {
			keyArr := strings.SplitN(dstPath, "#", 2)
			arrKey := keyArr[0]
			if gjson.Get(*s, arrKey).Exists() {
				return nil
			}
			*s, err = sjson.Set(*s, arrKey, "[]")
			if err != nil {
				err = errors.WithStack(err)
				return err
			}
			return nil
		}
		for index, val := range arr {
			path := strings.ReplaceAll(dstPath, "#", strconv.Itoa(index))
			*s, err = sjson.Set(*s, path, val)
			if err != nil {
				err = errors.WithStack(err)
				return err
			}
		}
		return nil
	}

	var realV interface{}
	realV = v //set default v with interface{} type
	if dstType == "" {
		*s, err = sjson.Set(*s, dstPath, realV)
		if err != nil {
			err = errors.WithStack(err)
			return err
		}
		return nil
	}
	strArr := make([]string, 0)
	intArr := make([]int, 0)
	int64Arr := make([]int64, 0)
	switch dstType {
	case reflect.String.String():
		realV = ""
	case reflect.Int.String(), "integer":
		realV = 0
	case reflect.Int64.String():
		realV = int64(0)
	case reflect.Uint.String():
		realV = uint(0)
	case reflect.Uint64.String():
		realV = uint64(0)
	case reflect.Float64.String():
		realV = float64(0)
	case reflect.Bool.String():
		realV = false
	case reflect.TypeOf(strArr).String():
		realV = strArr
	case reflect.TypeOf(intArr).String():
		realV = intArr
	case reflect.TypeOf(int64Arr).String():
		realV = int64Arr
	case reflect.Array.String(), reflect.Slice.String():
		realV = make([]interface{}, 0)
	}
	convertType(&realV, v)
	*s, err = sjson.Set(*s, dstPath, realV)
	if err != nil {
		err = errors.WithStack(err)
		return err
	}
	return nil
}
