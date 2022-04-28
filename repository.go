package templatemap

import (
	"bytes"
	"fmt"
	"io/fs"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	TPlSuffix      = ".tpl"
	REPOSITORY_KEY = "__repository"
)

type VolumeInterface interface {
	SetValue(key string, value interface{})
	GetValue(key string, value interface{}) (ok bool)
}

type Volume map[string]interface{}

func (v *Volume) init() {
	if v == nil {
		err := errors.Errorf("*Templatemap must init")
		panic(err)
	}
	if *v == nil {
		*v = Volume{} // 解决 data33 情况
	}
}

func (v *Volume) SetValue(key string, value interface{}) {
	v.init()
	(*v)[key] = value // todo 并发lock
}

func (v *Volume) GetValue(key string, value interface{}) bool {
	v.init()

	tmp, ok := getValue(v, key)
	if !ok {
		return false
	}
	ok = convertType(value, tmp)
	return ok

}

func getValue(v *Volume, key string) (interface{}, bool) {
	var mapKey string
	var jsonKey string
	var value interface{}
	var ok bool
	mapKey = key
	for {
		value, ok = (*v)[mapKey]
		if ok {
			break
		}
		lastIndex := strings.LastIndex(mapKey, ".")
		if lastIndex > -1 {
			mapKey = mapKey[:lastIndex]
			continue
		}
		break
	}
	if mapKey == key {
		return value, ok
	}
	// json key 获取值
	jsonKey = key[len(mapKey)+1:]
	jsonStr, ok := value.(string)
	if !ok {
		return nil, false
	}
	jsonValue, ok := getValueFromJson(jsonStr, jsonKey)
	return jsonValue, ok
}

func getValueFromJson(jsonStr string, jsonKey string) (interface{}, bool) {
	if jsonStr == "" {
		return nil, false
	}
	if !gjson.Valid(jsonStr) {
		err := errors.Errorf(`json str inValid %s`, jsonStr)
		panic(err)
	}
	key := jsonKey
	value := gjson.Result{}
	for {
		value = gjson.Get(jsonStr, key)
		if value.Exists() {
			break
		}
		lastIndex := strings.LastIndex(key, ".")
		if lastIndex > -1 {
			key = key[:lastIndex]
			continue
		}
		break
	}
	if jsonKey == key {
		return value.Value(), value.Exists()
	}

	return getValueFromJson(value.Str, jsonKey[len(key)+1:])
}

func convertType(dst interface{}, src interface{}) bool {
	if src == nil || dst == nil {
		return false
	}
	rv := reflect.ValueOf(dst)
	if rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Interface { // value 为 interface  指针时，使用 rv.Elem()
		rv = rv.Elem()
	}
	rvT := rv.Type()
	rTmp := reflect.ValueOf(src)
	ok := rTmp.CanConvert(rvT)
	if !ok {
		return false
	}
	val := rTmp.Convert(rvT)
	if rv.CanSet() {
		rv.Set(val)
		return true
	}

	if rv.Kind() == reflect.Ptr && rv.Elem().CanSet() {
		rv.Elem().Set(val.Elem())
		return true
	}
	panic("can not get value")
}

type ExecproviderInterface interface {
	Exec(identifier string, s string) (string, error)
}

type ExecProviderFunc func(identifier string, s string) (string, error)

func (f ExecProviderFunc) Exec(identifier string, s string) (string, error) {
	// 调用f函数本体
	return f(identifier, s)
}

type RepositoryInterface interface {
	AddTemplateByDir(dir string) (addTplNames []string)
	AddTemplateByFS(fsys fs.FS, root string) (addTplNames []string)
	AddTemplateByStr(name string, s string) (addTplNames []string)
	GetTemplate() *template.Template
	ExecuteTemplate(name string, volume VolumeInterface) (out string, err error)
	TemplateExists(name string) bool
	RegisterProvider(provider ExecproviderInterface, tplNames ...string)
	GetProvider(tplName string) (ExecproviderInterface, bool)
}

type repository struct {
	template     *template.Template
	providerPool map[string]ExecproviderInterface
}

func NewRepository() RepositoryInterface {
	r := &repository{
		template:     newTemplate(),
		providerPool: make(map[string]ExecproviderInterface),
	}
	return r
}

func newTemplate() *template.Template {
	return template.New("").Funcs(CoreFuncMap).Funcs(TemplatefuncMap).Funcs(sprig.TxtFuncMap())
}

func (r *repository) RegisterProvider(providor ExecproviderInterface, tplNames ...string) {
	for _, tplName := range tplNames {
		r.providerPool[tplName] = providor
	}
}

func (r *repository) GetProvider(tplName string) (ExecproviderInterface, bool) {
	provider, ok := r.providerPool[tplName]
	return provider, ok
}

func (r *repository) GetTemplate() *template.Template {
	return r.template
}

func (r *repository) AddTemplateByDir(dir string) []string {

	patten := fmt.Sprintf("%s/**%s", strings.TrimRight(dir, "/"), TPlSuffix)
	allFileList, err := GlobDirectory(dir, patten)
	if err != nil {
		err = errors.WithStack(err)
		panic(err)
	}

	r.template = template.Must(r.template.ParseFiles(allFileList...)) // 追加
	tmp := template.Must(newTemplate().ParseFiles(allFileList...))
	out := getTemplateNames(tmp)
	return out
}

func (r *repository) AddTemplateByFS(fsys fs.FS, root string) []string {
	patten := fmt.Sprintf("%s/**%s", strings.TrimRight(root, "/"), TPlSuffix)
	allFileList, err := GlobFS(fsys, patten)
	if err != nil {
		err = errors.WithStack(err)
		panic(err)
	}
	r.template = template.Must(parseFiles(r.template, readFileFS(fsys), allFileList...)) // 追加
	tmp := template.Must(parseFiles(newTemplate(), readFileFS(fsys), allFileList...))
	out := getTemplateNames(tmp)
	return out
}

func (r *repository) AddTemplateByStr(name string, s string) []string {
	var tmpl *template.Template
	if name == r.template.Name() {
		tmpl = r.template
	} else {
		tmpl = r.template.New(name)
	}
	template.Must(tmpl.Parse(s)) // 追加

	tmp := template.Must(newTemplate().Parse(s))
	out := getTemplateNames(tmp)
	return out
}

func (r *repository) ExecuteTemplate(name string, volume VolumeInterface) (string, error) {
	if volume == nil {
		volume = &Volume{}
	} else {
		volumeR := reflect.ValueOf(volume)
		if volumeR.IsNil() {
			err := errors.Errorf("%#v must not nil", volumeR)
			return "", err

		}
	}
	volume.SetValue(REPOSITORY_KEY, r) // 将模板传入，方便在模板中执行模板
	var b bytes.Buffer
	err := r.template.ExecuteTemplate(&b, name, volume)
	if err != nil {
		err = errors.WithStack(err)
		return "", err
	}
	out := b.String()
	return out, nil

}

func (r *repository) TemplateExists(name string) bool {
	t := r.template.Lookup(name)
	return t != nil
}
