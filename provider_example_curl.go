package templatemap

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	EOF                  = "\n"
	WINDOW_EOF           = "\r\n"
	HTTP_HEAD_BODY_DELIM = EOF + EOF
)

var CURL_TIMEOUT = 30 * time.Millisecond

type RequestData struct {
	URL     string         `json:"url"`
	Method  string         `json:"method"`
	Header  http.Header    `json:"header"`
	Cookies []*http.Cookie `json:"cookies"`
	Body    string         `json:"body"`
}
type ResponseData struct {
	HttpStatus  string         `json:"httpStatus"`
	Header      http.Header    `json:"header"`
	Cookies     []*http.Cookie `json:"cookies"`
	Body        string         `json:"body"`
	RequestData *RequestData   `json:"requestData"`
}

func InitHTTPClient(p *CURLExecProvider) *http.Client {

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, // 连接超时时间
			KeepAlive: 30 * time.Second, // 连接保持超时时间
		}).DialContext,
		MaxIdleConns:        2000,             // 最大连接数,默认0无穷大
		MaxIdleConnsPerHost: 2000,             // 对每个host的最大连接数量(MaxIdleConnsPerHost<=MaxIdleConns)
		IdleConnTimeout:     90 * time.Second, // 多长时间未使用自动关闭连
	}
	if p.Config.Proxy != "" {
		proxy, err := url.Parse(p.Config.Proxy)
		if err != nil {
			panic(err)
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	httpClient := &http.Client{
		Transport: transport,
	}
	return httpClient
}

type CURLExecProviderConfig struct {
	Proxy    string        `json:"proxy"`
	LogLevel string        `json:"logLevel"`
	Timeout  time.Duration `json:"timeout"`
}

type CURLExecProvider struct {
	Config     CURLExecProviderConfig
	InitClient func() *http.Client
	client     *http.Client
	clinetOnce sync.Once
}

func (p *CURLExecProvider) Exec(identifier string, s string) (string, error) {
	return CURlProvider(p, s)
}

func (p *CURLExecProvider) Begin() (tx interface{}, err error) {
	return nil, nil
}

func (p *CURLExecProvider) Rollback(tx interface{}) (err error) {
	return nil
}

func (p *CURLExecProvider) Commit(tx interface{}) (err error) {
	return nil
}

// GetDb is a signal DB
func (p *CURLExecProvider) GetClient() *http.Client {
	if p.client == nil {
		if p.InitClient == nil {
			p.InitClient = func() *http.Client { return InitHTTPClient(p) }
		}
		p.clinetOnce.Do(func() {
			p.client = p.InitClient()

		})
	}
	return p.client
}

func CURlProvider(p *CURLExecProvider, httpRaw string) (string, error) {
	reqReader, err := ReadRequest(httpRaw)
	if err != nil {
		return "", err
	}
	reqData, err := Request2RequestData(reqReader)
	if err != nil {
		return "", err
	}
	timeout := p.Config.Timeout
	if timeout == 0 {
		timeout = 30
	}
	timeout = timeout * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, reqData.Method, reqData.URL, bytes.NewReader([]byte(reqData.Body)))
	if err != nil {
		return "", err
	}

	for k, vArr := range reqData.Header {
		for _, v := range vArr {
			req.Header.Add(k, v)
		}
	}

	rsp, err := p.GetClient().Do(req)
	if err != nil {
		return "", err
	}

	defer rsp.Body.Close()
	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	if err != nil {
		msg := fmt.Sprintf("response httpstatus:%d, body: %s", rsp.StatusCode, string(b))
		err = errors.WithMessage(err, msg)
		return "", err
	}
	if rsp.StatusCode != http.StatusOK {
		err := errors.Errorf("response httpstatus:%d, body: %s", rsp.StatusCode, string(b))
		return "", err
	}

	rspData := ResponseData{
		HttpStatus:  strconv.Itoa(rsp.StatusCode),
		Header:      rsp.Header,
		Cookies:     rsp.Cookies(),
		RequestData: reqData,
	}
	rspData.Body = string(b)
	jsonByte, err := json.Marshal(rspData)
	if err != nil {
		return "", err
	}
	out := string(jsonByte)

	return out, nil
}

func ReadRequest(httpRaw string) (req *http.Request, err error) {
	httpRaw = TrimSpaces(httpRaw) // （删除前后空格，对于没有body 内容的请求，后面再加上必要的换行）
	if httpRaw == "" {
		err = errors.Errorf("http raw not allow empty")
		return nil, err
	}
	httpRaw = strings.ReplaceAll(httpRaw, "\r\n", "\n") // 统一换行符
	// 插入body长度头部信息
	bodyIndex := strings.Index(httpRaw, HTTP_HEAD_BODY_DELIM)
	formatHttpRaw := httpRaw
	if bodyIndex > 0 {
		headerRaw := strings.TrimSpace(httpRaw[:bodyIndex])
		bodyRaw := httpRaw[bodyIndex+len(HTTP_HEAD_BODY_DELIM):]
		bodyLen := len(bodyRaw)
		formatHttpRaw = fmt.Sprintf("%s%sContent-Length: %d%s%s", headerRaw, EOF, bodyLen, HTTP_HEAD_BODY_DELIM, bodyRaw)
	} else {
		// 如果没有请求体，则原始字符后面必须保留一个换行符
		formatHttpRaw = fmt.Sprintf("%s%s", formatHttpRaw, HTTP_HEAD_BODY_DELIM)
	}

	buf := bufio.NewReader(strings.NewReader(formatHttpRaw))
	req, err = http.ReadRequest(buf)
	if err != nil {
		return
	}
	if req.URL.Scheme == "" {
		queryPre := ""
		if req.URL.RawQuery != "" {
			queryPre = "?"
		}
		req.RequestURI = fmt.Sprintf("http://%s%s%s%s", req.Host, req.URL.Path, queryPre, req.URL.RawQuery)
	}

	return
}

func Request2RequestData(req *http.Request) (requestData *RequestData, err error) {
	requestData = &RequestData{}
	bodyByte, err := io.ReadAll(req.Body)
	if err != nil {
		return
	}
	req.Header.Del("Content-Length")
	requestData = &RequestData{
		URL:     req.URL.String(),
		Method:  req.Method,
		Header:  req.Header,
		Cookies: req.Cookies(),
		Body:    string(bodyByte),
	}

	return
}
