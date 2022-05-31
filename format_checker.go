package templatemap

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/xeipuuv/gojsonschema"
)

var (
	rxPhone    = regexp.MustCompile(`/^1[3456789]\d{9}$/`)
	rxIdCard   = regexp.MustCompile(`/^[1-9]\d{5}(18|19|([23]\d))\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$/`)
	rxIdCard1  = regexp.MustCompile(`/^[1-9]\d{5}\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}$/`)
	rxPostCode = regexp.MustCompile(`/^[1-9]{1}(\d+){5}$/`)
)

func RegisterFormatChecker() {
	gojsonschema.FormatCheckers.Add("number", NumberFormatChecker{}) // 数字格式验证
	gojsonschema.FormatCheckers.Add("phone", NumberFormatChecker{})  // 数字格式验证
}

type NumberFormatChecker struct{}

// IsFormat checks if input is a correctly formatted number string
func (f NumberFormatChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}
	_, err := strconv.ParseFloat(asString, 64)
	return err == nil
}

type PhoneFormatChecker struct{}

// IsFormat checks if input is a correctly formatted phone string
func (f PhoneFormatChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}
	out := rxPhone.MatchString(asString)
	return out
}

type IDCardFormatChecker struct{}

// IsFormat checks if input is a correctly formatted IDCard string
func (f IDCardFormatChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}
	out := rxIdCard.MatchString(asString) || rxIdCard1.MatchString(asString)
	return out
}

type PostCodeFormatChecker struct{}

// IsFormat checks if input is a correctly formatted postcode string
func (f PostCodeFormatChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}
	out := rxPostCode.MatchString(asString)
	return out
}

//  数据库验证(这个需要调用时重新注册，方便更新TplName 等数据)
type ValidDBChecker struct {
	Repository RepositoryInterface
	TplName    string
	Volume     VolumeInterface
}

func (f *ValidDBChecker) Name() string {
	return "DBValidate"
}

func (f *ValidDBChecker) IsFormat(input interface{}) bool {

	err := f.Repository.ExecuteTemplate(f.TplName, f.Volume)
	if err != nil {
		panic(err)
	}
	key := fmt.Sprintf("%sOut.ok", f.TplName)
	var ok bool
	f.Volume.GetValue(key, &ok)
	return ok
}

func (f *ValidDBChecker) Msg() string {
	msgKey := fmt.Sprintf("%sOut.msg", f.TplName)
	var msg string
	f.Volume.GetValue(msgKey, &msg)
	return msg
}
