package errno

import (
	"encoding/json"
	"net/http"
)

var _ Error = (*err)(nil)

type Error interface {
	// i 为了避免被其他包实现
	i()
	// WithData 设置成功时返回的数据
	WithData(data interface{}) Error
	// GetBusinessCode 获取 Business Code
	GetBusinessCode() int
	// GetHttpCode 获取 HTTP Code
	GetHttpCode() int
	// GetMsg 获取 Msg
	GetMsg() string
	// ToString 返回 JSON 格式的错误详情
	ToString() string
}

type err struct {
	HttpCode     int         `json:"-"`    // HTTP Code
	BusinessCode int         `json:"code"` // Business Code
	Msg          string      `json:"msg"`  // 错误描述
	Data         interface{} `json:"data"` // 成功时返回的数据
}

func NewError(businessCode int, msg string) Error {
	return &err{
		HttpCode:     http.StatusOK,
		BusinessCode: businessCode,
		Msg:          msg,
		Data:         nil,
	}
}
func SysemError(httpCode, businessCode int, msg string) Error {
	return &err{
		HttpCode:     httpCode,
		BusinessCode: businessCode,
		Msg:          msg,
		Data:         nil,
	}
}

func (e *err) i() {}

func (e *err) WithData(data interface{}) Error {
	e.Data = data
	return e
}

func (e *err) GetHttpCode() int {
	return e.HttpCode
}

func (e *err) GetBusinessCode() int {
	return e.BusinessCode
}

func (e *err) GetMsg() string {
	return e.Msg
}

// ToString 返回 JSON 格式的错误详情
func (e *err) ToString() string {
	err := &struct {
		HttpCode     int         `json:"http_code"`
		BusinessCode int         `json:"business_code"`
		Msg          string      `json:"msg"`
		Data         interface{} `json:"data"`
	}{
		HttpCode:     e.HttpCode,
		BusinessCode: e.BusinessCode,
		Msg:          e.Msg,
		Data:         e.Data,
	}

	raw, _ := json.Marshal(err)
	return string(raw)
}
