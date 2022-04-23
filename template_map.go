package templatemap

import (
	"bytes"
	"fmt"
	"io/fs"
	"reflect"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

const (
	TPlSuffix      = ".tpl"
	REPOSITORY_KEY = "__repository"
)

type VolumeInterface interface {
	SetValue(key string, value interface{})
	GetValue(key string, value interface{}) (ok bool)
	GetDynamicValus() (values map[string]interface{})
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

	tmp, ok := (*v)[key]
	if !ok {
		return false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Interface { // value 为 interface  指针时，使用 rv.Elem()
		rv = rv.Elem()
	}
	rvT := rv.Type()
	rTmp := reflect.ValueOf(tmp)
	fmt.Println(rvT.Name())
	ok = rTmp.CanConvert(rvT)
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
	panic("can not set value")
}

func (v *Volume) GetDynamicValus() (values map[string]interface{}) {
	v.init()
	return *v
}

type TempalteMetaInterface interface {
	TplName() string
	TplType() string // 返回 TPL_DEFINE_TYPE 类型，方便后续根据类型获取资源(db、curl) 自动获取数据
}
type TemplateMeta struct {
	tplName string
	tplType string
}

func (tm *TemplateMeta) TplName() string {
	return tm.tplName
}
func (tm *TemplateMeta) TplType() string {
	return tm.tplType
}

type RepositoryInterface interface {
	AddTemplateByDir(dir string) (err error)
	AddTemplateByFS(fsys fs.FS, root string) (err error)
	AddTemplateByStr(name string, s string) (err error)
	GetTemplate() *template.Template
	ExecuteTemplate(name string, volume VolumeInterface) (string, error)
}

type repository struct {
	template *template.Template
}

func NewRepository(funcMap template.FuncMap) RepositoryInterface {
	return &repository{
		template: template.New("").Funcs(funcMap),
	}
}

func (r *repository) GetTemplate() *template.Template {
	return r.template
}
func (r *repository) Delims(left string, right string) {
	r.template = r.template.Delims(left, right)
}

func (r *repository) AddTemplateByDir(dir string) (err error) {

	patten := fmt.Sprintf("%s/**%s", strings.TrimRight(dir, "/"), TPlSuffix)
	allFileList, err := GlobDirectory(dir, patten)
	if err != nil {
		err = errors.WithStack(err)
		return err
	}
	r.template, err = r.template.ParseFiles(allFileList...)
	if err != nil {
		return err
	}
	return
}

func (r *repository) AddTemplateByFS(fsys fs.FS, root string) (err error) {
	patten := fmt.Sprintf("%s/**%s", strings.TrimRight(root, "/"), TPlSuffix)
	allFileList, err := GlobFS(fsys, patten)
	if err != nil {
		err = errors.WithStack(err)
		return err
	}
	r.template, err = parseFiles(r.template, readFileFS(fsys), allFileList...)
	if err != nil {
		return err
	}
	return
}

func (r *repository) AddTemplateByStr(name string, s string) (err error) {
	var tmpl *template.Template
	if name == r.template.Name() {
		tmpl = r.template
	} else {
		tmpl = r.template.New(name)
	}
	_, err = tmpl.Parse(s)
	if err != nil {
		return err
	}
	return
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
