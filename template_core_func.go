package templatemap

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/suifengpiao14/templatemap/provider"
	"github.com/suifengpiao14/templatemap/util"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	gormLogger "gorm.io/gorm/logger"
)

var CoreFuncMap = template.FuncMap{
	"executeTemplate":                  ExecuteTemplate,
	"setValue":                         SetValue,
	"panic":                            Panic,
	"getValue":                         GetValue,
	"getSetValue":                      GetSetValue,
	"getSetValueInt":                   GetSetValueInt,
	"getSetNumber":                     GetSetValueNumber,
	"getSetValueNumber":                GetSetValueNumber,
	"getSetValueNumberWithOutEmptyStr": GetSetValueNumberWithOutEmptyStr,
	"getSetColumn2Row":                 GetSetColumn2Row,
	"getSetRow2Column":                 GetSetRow2Column,
	"toSQL":                            ToSQL,
	"exec":                             Exec,
	"execBinTpl":                       ExecBinTpl,
	"execSQLTpl":                       ExecSQLTpl,
	"execCURLTpl":                      ExecCURLTpl,
	"gjsonGet":                         gjson.Get,
	"sjsonSet":                         sjson.Set,
	"sjsonSetRaw":                      sjson.SetRaw,
	"transfer":                         Transfer,
	"DBValidate":                       DBValidate,
	"dbValidate":                       DBValidate,
	"toBool":                           ToBool,
	"getSource":                        GetSource,
	"listPadIndex":                     ListPadIndex, //生成指定长度的整型数组，变相在模板中实现for
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

// ExecuteTemplate 模板中调用模板
func ExecuteTemplate(volume VolumeInterface, name string) string {
	var r = getRepositoryFromVolume(volume)
	out, err := r.ExecuteTemplate(name, volume)
	if err != nil {
		panic(err) // ExecuteTemplate 函数可能嵌套很多层，抛出错误值后有可能被当成正常值处理，所以此处直接panic 退出，保留原始错误输出
	}
	return out
}

func SetValue(volume VolumeInterface, key string, value interface{}) string { // SetValue 返回空字符，不对模板产生新输出
	volume.SetValue(key, value)
	return ""
}

func ToBool(v interface{}) bool {
	var ok bool
	switch v := v.(type) {
	case bool:
		ok = v
	case string:
		ok = v != "" && v != "false" && v != "0"
	case int:
		ok = v > 0
	case int64:
		ok = v > 0
	}
	return ok
}

func ListPadIndex(size int) (out []int) {
	out = make([]int, size)
	return
}

func DBValidate(volume VolumeInterface, ok bool, msg string) string {
	value := fmt.Sprintf(`{"ok":%v,"msg":"%s"}`, ok, msg)
	return value
}

func Panic(httpCode string, businessCode string, msg string) string {
	err := errors.Errorf("%s#%s#%s", httpCode, businessCode, msg)
	panic(err)
}

func GetValue(volume VolumeInterface, key string) interface{} {
	var value interface{}
	volume.GetValue(key, &value)
	return value
}

func GetSetValue(volume VolumeInterface, setKey string, getKey string) string {
	v := GetValue(volume, getKey)
	volume.SetValue(setKey, v)
	return ""
}
func GetSetValueInt(volume VolumeInterface, setKey string, getKey string) string {
	var v int
	volume.GetValue(getKey, &v)
	volume.SetValue(setKey, v)
	return ""
}

func GetSetValueNumberWithOutEmptyStr(volume VolumeInterface, setKey string, getKey string) string {
	var v string
	volume.GetValue(getKey, &v)
	v = strings.TrimSpace(v)
	if v == "" {
		err := errors.Errorf("key(%s) required number format,got empty", getKey)
		panic(err)
	}
	if strings.Contains(v, ".") {
		oFloat, err := strconv.ParseFloat(v, 64)
		if err != nil {
			err = errors.WithStack(err)
			panic(err)
		}
		volume.SetValue(setKey, oFloat)
		return ""
	}
	oInt, err := strconv.Atoi(v)
	if err != nil {
		err = errors.WithStack(err)
		panic(err)
	}
	volume.SetValue(setKey, oInt)
	return ""
}

func GetSetValueNumber(volume VolumeInterface, setKey string, getKey string) string {
	var v string
	volume.GetValue(getKey, &v)
	if v == "" {
		volume.SetValue(setKey, 0)
		return ""
	}
	return GetSetValueNumberWithOutEmptyStr(volume, setKey, getKey)
}

func GetSetColumn2Row(volume VolumeInterface, key string) string {
	var v string
	volume.GetValue(key, &v) // 多级key,返回的为map类型,无法转换为string
	if v == "" {
		return ""
	}
	out := util.Column2Row(v)
	volume.SetValue(key, out)
	return ""
}
func GetSetRow2Column(volume VolumeInterface, key string) string {
	var v string
	volume.GetValue(key, &v)
	if v == "" {
		return ""
	}
	out := util.Row2Column(v)
	volume.SetValue(key, out)
	return ""
}
func GetSource(volume VolumeInterface, tplName string) (source interface{}) {
	provider := GetProvider(volume, tplName)
	return provider.GetSource()
}
func GetProvider(volume VolumeInterface, tplName string) provider.ExecproviderInterface {
	var r = getRepositoryFromVolume(volume)
	meta, ok := r.GetMeta(tplName)
	if !ok {
		err := errors.Errorf("templatemap.Exec: not found meta  by template name : %s", tplName)
		panic(err)
	}

	execProvider := meta.ExecProvider
	if execProvider == nil {
		err := errors.Errorf("meta:%v provider must be set", meta)
		panic(err)
	}
	return execProvider
}

func Exec(volume VolumeInterface, tplName string, s string) string {
	provider := GetProvider(volume, tplName)
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

func ExecBinTpl(volume VolumeInterface, templateName string) error {
	tplOut := ExecuteTemplate(volume, templateName)
	out := Exec(volume, templateName, tplOut)
	storeKey := fmt.Sprintf("%sOut", templateName)
	volume.SetValue(storeKey, out)
	return nil
}

func ExecSQLTpl(volume VolumeInterface, templateName string) string {
	//{{executeTemplate . "Paginate"|toSQL . | exec . "docapi_db2"|setValue . }}
	tplOut := ExecuteTemplate(volume, templateName)
	tplOut = util.StandardizeSpaces(tplOut)
	if tplOut == "" {
		err := errors.Errorf("sql template :%s return empty sql", templateName)
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

// FormatJson 根据json schema 格式化 json数据，填充默认值，容许为空时，设置类型初始化值等
func FormatJson(jsonStr string, jsonschema string) (string, error) {
	out := jsonStr
	var err error
	schema := NewJsonSchema(jsonschema)
	schema.SetSrcAsDst()                       // 自动填充src，方便统一调用函数
	transferPaths := schema.GetTransferPaths() // 此处只是用dst 即可
	for _, transferPath := range transferPaths {
		dstResult := gjson.Get(out, transferPath.Dst)
		if !dstResult.Exists() {
			if transferPath.Default == "__nil__" {
				continue
			}
			if transferPath.Default != nil {
				out, err = sjson.Set(out, transferPath.Dst, transferPath.Default)
				if err != nil {
					return "", err
				}
			}
			// 设置类型初始化值
			var v interface{}
			switch transferPath.DstType {
			case "string":
				v = ""
			case "int", "integer", "number":
				v = 0
			case "array":
				v = make([]interface{}, 0)
			case "float":
				v = 0.0
			}
			out, err = sjson.Set(out, transferPath.Dst, v)
			if err != nil {
				err = errors.WithStack(err)
				return "", err
			}
		}
	}
	return out, nil
}

func TransferDataFromVolume(volume VolumeInterface, transferPaths TransferPaths) (string, error) {
	out := ""
	var err error
	parentTransferPaths := TransferPaths{}
	for _, tp := range transferPaths {
		var v interface{}
		var dst = tp.Dst
		var dstType = tp.DstType
		ok := volume.GetValue(tp.Src, &v)
		if !ok {
			optionalTp := tp.GetOptionalTransferPath()
			if optionalTp == nil {
				err := errors.Errorf("not found %s data from volume %#v", tp.Src, volume)
				return "", err
			}

			if tp.Default == nil && tp.Parent != nil {
				parentTransferPaths = append(parentTransferPaths, tp.Parent)
				continue // 没有默认值，则直接跳过
			}
			v = tp.Default
		}
		err = Add2json(&out, dst, dstType, v)
		if err != nil {
			return "", err
		}
	}

	// 补充parent, 解决子对象全部为空情况

	parentTransferPaths = parentTransferPaths.UniqueItems()
	for _, tp := range parentTransferPaths {
		AddParent2Json(&out, *tp)
	}

	return out, nil
}

// AddParent2Json 填充父类元素，解决子元素全部为空情况
func AddParent2Json(s *string, tp TransferPath) error {
	var err error
	typeValueArr, _ := tp.Schema.MultiType()
	typeValue := ""
	if len(typeValueArr) > 0 {
		typeValue = strings.ToLower(typeValueArr[0])

	}

	if tp.Schema.IsRoot() && *s == "" {
		switch typeValue {
		case "object":
			*s = "{}"
		case "array":
			*s = "[]"
		}
		return nil
	}
	dstLen := len(tp.Dst)
	if dstLen < 1 {
		err = errors.Errorf("TransferPath.dst is empty :%#v", tp)
		panic(err)
	}
	if tp.Dst[dstLen-1:] == "#" {
		if tp.Parent == nil {
			return nil
		}
		return AddParent2Json(s, *tp.Parent)
	}
	if gjson.Get(*s, tp.Dst).Exists() {
		return nil
	}
	var v interface{}
	v = tp.Default
	if v == nil {

		switch typeValue {
		case "":
			if tp.Parent != nil {
				return AddParent2Json(s, *tp.Parent)
			}
		case "object":
			v = map[string]interface{}{}
		case "array":
			v = make([]string, 0)

		}
	}
	*s, err = sjson.Set(*s, tp.Dst, v)
	if err != nil {
		return err
	}
	return nil
}

// Add2json 数据转换(将go数据写入到json字符串中)
func Add2json(s *string, dstPath string, dstType string, v interface{}) error {
	var err error
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
