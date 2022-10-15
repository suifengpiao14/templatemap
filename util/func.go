package util

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// 拷贝template 包helper 方法
func ReadFileFS(fsys fs.FS) func(string) (string, []byte, error) {
	return func(file string) (name string, b []byte, err error) {
		name = path.Base(file)
		b, err = fs.ReadFile(fsys, file)
		return
	}
}

// parseFiles is the helper for the method and function. If the argument
// template is nil, it is created from the first file.
func ParseFiles(t *template.Template, readFile func(string) (string, []byte, error), filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("template: no files named in call to ParseFiles")
	}
	for _, filename := range filenames {
		name, b, err := readFile(filename)
		if err != nil {
			return nil, err
		}
		s := string(b)
		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.
		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func GlobDirectory(dir string, pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		pattern = fmt.Sprintf("%s/*%s", dir, pattern)
		return filepath.Glob(pattern)
	}
	var matches []string
	reg := getPattern(pattern)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			err := errors.Errorf("dir:%s filepath.Walk info is nil", dir)
			return err
		}
		if !info.IsDir() {
			path = strings.ReplaceAll(path, "\\", "/")
			if reg.MatchString(path) {
				matches = append(matches, path)
			}
		}
		return nil
	})
	return matches, err
}

func GlobFS(fsys fs.FS, pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		return fs.Glob(fsys, pattern)
	}
	var matches []string
	reg := getPattern(pattern)
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			if reg.MatchString(path) {
				matches = append(matches, path)
			}
		}
		return nil
	})
	return matches, err
}

func getPattern(pattern string) *regexp.Regexp {
	regStr := strings.TrimLeft(pattern, ".")
	regStr = strings.ReplaceAll(regStr, ".", "\\.")
	regStr = strings.ReplaceAll(regStr, "**", ".*")
	reg := regexp.MustCompile(regStr)
	return reg
}

// TrimSpaces  去除开头结尾的非有效字符
func TrimSpaces(s string) string {
	return strings.Trim(s, "\r\n\t\v\f ")
}

func StandardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func GetTemplateNames(t *template.Template) []string {
	out := make([]string, 0)
	for _, tpl := range t.Templates() {
		name := tpl.Name()
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func Validate(input string, jsonLoader gojsonschema.JSONLoader) (err error) {
	if input == "" {
		jsonschema, err := jsonLoader.LoadJSON()
		if err != nil {
			return err
		}
		jsonMap, ok := jsonschema.(map[string]interface{})
		if !ok {
			err = errors.Errorf("can not convert jsonLoader.LoadJSON() to map[string]interface{}")
			return err
		}
		typ, ok := jsonMap["type"]
		if !ok {
			err = errors.Errorf("jsonschema missing property type :%v", jsonschema)
			return err
		}
		typStr, ok := typ.(string)
		if !ok {
			err = errors.Errorf("can not convert  jsonschema type to string :%v", typ)
			return err

		}
		switch strings.ToLower(typStr) {
		case "object":
			input = "{}"
		case "array":
			input = "[]"
		default:
			err = errors.Errorf("invalid jsonschema type:%v", typStr)
			return err
		}

	}
	documentLoader := gojsonschema.NewStringLoader(input)
	result, err := gojsonschema.Validate(jsonLoader, documentLoader)
	if err != nil {
		return err
	}
	if result.Valid() {
		return nil
	}

	msgArr := make([]string, 0)
	for _, resultError := range result.Errors() {
		msgArr = append(msgArr, resultError.String())
	}
	err = errors.Errorf("input args validate errors: %s", strings.Join(msgArr, ","))
	return err
}

// Column2Row 列数据(二维数组中一维为列对象，二维为值数组)转行数据(二维数组中，一维为行索引，二维为行对象) gjson 获取数据时，会将行数据，转换成列数据，此时需要调用该函数再转换为行数据
func Column2Row(jsonStr string) (out string) {
	arr := make(map[string][]interface{}, 0)
	err := json.Unmarshal([]byte(jsonStr), &arr)
	if err != nil {
		panic(err)
	}
	var outArr []map[string]interface{}
	initOutArr := true
	for key, column := range arr {
		if initOutArr { //init array
			outArr = make([]map[string]interface{}, len(column))
			initOutArr = false
		}

		for i, val := range column {
			if outArr[i] == nil { //init map
				outArr[i] = make(map[string]interface{})
			}
			outArr[i][key] = val
		}
	}
	b, err := json.Marshal(outArr)
	if err != nil {
		panic(err)
	}
	out = string(b)
	return out
}

// Row2Column 行数据(二维数组中，一维为行索引，二维为行对象)转 列数据(二维数组中一维为列对象，二维为值数组) ，有时json数据的key和结构体的json key（结构体由第三方包定义，不可修改tag） 不一致，但是存在对应关系，需要构造结构体内json key数据，此时将行数据改成列数据，方便计算
func Row2Column(jsonStr string) (out string) {
	arr := make([]map[string]interface{}, 0)
	err := json.Unmarshal([]byte(jsonStr), &arr)
	if err != nil {
		panic(err)
	}
	outMap := make(map[string][]interface{}, 0)
	arrLen := len(arr)
	for i, row := range arr {
		for key, val := range row {
			if outMap[key] == nil {
				outMap[key] = make([]interface{}, arrLen)
			}
			outMap[key][i] = val
		}
	}
	b, err := json.Marshal(outMap)
	if err != nil {
		panic(err)
	}
	out = string(b)
	return out
}
