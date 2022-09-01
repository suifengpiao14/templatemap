package provider

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	PROVIDER_SQL      = "SQL"
	PROVIDER_CURL     = "CURL"
	PROVIDER_BIN      = "BIN"
	PROVIDER_REDIS    = "REDIS"
	PROVIDER_RABBITMQ = "RABBITMQ"
)

type ExecproviderInterface interface {
	Exec(identifier string, s string) (string, error)
	GetSource() (source interface{})
}

type ExecProviderFunc func(identifier string, s string) (string, error)

func (f ExecProviderFunc) Exec(identifier string, s string) (string, error) {
	// 调用f函数本体
	return f(identifier, s)
}

//MakeExecProvider 根据名称，获取exec 执行器，后续改成注册执行器方式
func MakeExecProvider(identifier string, configJson string) (execProvider ExecproviderInterface, err error) {

	switch identifier {
	case PROVIDER_SQL:
		var config DBExecProviderConfig
		if configJson != "" {
			err = json.Unmarshal([]byte(configJson), &config)
			if err != nil {
				return nil, err
			}
		}
		execProvider = &DBExecProvider{
			Config: config,
		}
	case PROVIDER_CURL:
		var config CURLExecProviderConfig
		if configJson != "" {
			err = json.Unmarshal([]byte(configJson), &config)
			if err != nil {
				return nil, err
			}
		}
		execProvider = &CURLExecProvider{
			Config: config,
		}
	case PROVIDER_BIN:
		execProvider = &BinExecProvider{}
	default:
		err = errors.Errorf("not suport source type :%s", identifier)
		return nil, err
	}
	return execProvider, nil
}
