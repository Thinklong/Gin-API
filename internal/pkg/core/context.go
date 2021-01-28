package core

import (
	"bytes"
	errcode "gin-api/internal/api/code"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gin-api/internal/pkg/trace"
	"gin-api/pkg/errno"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

type HandlerFunc func(c Context)

type Trace = trace.T

const (
	_Alias          = "_alias_"
	_TraceName      = "_trace_"
	_LoggerName     = "_logger_"
	_BodyName       = "_body_"
	_PayloadName    = "_payload_"
	_UserID         = "_user_id_"
	_UserName       = "_user_name_"
	_Lang           = "_lang_"
	_AbortErrorName = "_abort_error_"
)

var contextPool = &sync.Pool{
	New: func() interface{} {
		return new(context)
	},
}

func newContext(ctx *gin.Context) Context {
	context := contextPool.Get().(*context)
	context.ctx = ctx
	return context
}

func releaseContext(ctx Context) {
	c := ctx.(*context)
	c.ctx = nil
	contextPool.Put(c)
}

var _ Context = (*context)(nil)

type Context interface {
	init()

	// ShouldBindQuery 反序列化querystring
	// tag: `form:"xxx"` (注：不要写成query)
	ShouldBindQuery(obj interface{}) error

	// ShouldBindPostForm 反序列化postform(querystring会被忽略)
	// tag: `form:"xxx"`
	ShouldBindPostForm(obj interface{}) error

	// ShouldBindForm 同时反序列化querystring和postform;
	// 当querystring和postform存在相同字段时，postform优先使用。
	// tag: `form:"xxx"`
	ShouldBindForm(obj interface{}) error

	// ShouldBindJSON 反序列化postjson
	// tag: `json:"xxx"`
	ShouldBindJSON(obj interface{}) error

	// ShouldBindURI 反序列化path参数(如路由路径为 /user/:name)
	// tag: `uri:"xxx"`
	ShouldBindURI(obj interface{}) error

	// Redirect 重定向
	Redirect(code int, location string)

	Trace() Trace
	setTrace(trace Trace)
	disableTrace()

	Logger() *zap.Logger
	setLogger(logger *zap.Logger)

	getPayload() errno.Error
	SetPayload(payload errno.Error)

	Success(data interface{})
	Fail(code int)

	Header() http.Header
	GetHeader(key string) string
	SetHeader(key, value string)

	UserID() int
	setUserID(userID int)

	UserName() string
	setUserName(userName string)

	Lang() string

	AbortWithError(err errno.Error)
	abortError() errno.Error

	Alias() string
	setAlias(path string)

	RawData() []byte
	Method() string
	Host() string
	Path() string
	URI() string

	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}

type context struct {
	ctx *gin.Context
}

func (c *context) init() {
	body, err := c.ctx.GetRawData()
	if err != nil {
		panic(err)
	}

	if lang := c.GetHeader("lang"); lang != "" {
		c.ctx.Set(_Lang, lang)
	}

	c.ctx.Set(_BodyName, body)                                   // cache body是为了trace使用
	c.ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // re-construct req body
}

// ShouldBindQuery 反序列化querystring
// tag: `form:"xxx"` (注：不要写成query)
func (c *context) ShouldBindQuery(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.Query)
}

// ShouldBindPostForm 反序列化postform(querystring会被忽略)
// tag: `form:"xxx"`
func (c *context) ShouldBindPostForm(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.FormPost)
}

// ShouldBindForm 同时反序列化querystring和postform;
// 当querystring和postform存在相同字段时，postform优先使用。
// tag: `form:"xxx"`
func (c *context) ShouldBindForm(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.Form)
}

// ShouldBindJSON 反序列化postjson
// tag: `json:"xxx"`
func (c *context) ShouldBindJSON(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindURI 反序列化path参数(如路由路径为 /user/:name)
// tag: `uri:"xxx"`
func (c *context) ShouldBindURI(obj interface{}) error {
	return c.ctx.ShouldBindUri(obj)
}

// Redirect 重定向
func (c *context) Redirect(code int, location string) {
	c.ctx.Redirect(code, location)
}

func (c *context) Trace() Trace {
	t, ok := c.ctx.Get(_TraceName)
	if !ok || t == nil {
		return nil
	}

	return t.(Trace)
}

func (c *context) setTrace(trace Trace) {
	c.ctx.Set(_TraceName, trace)
}

func (c *context) disableTrace() {
	c.setTrace(nil)
}

func (c *context) Logger() *zap.Logger {
	logger, ok := c.ctx.Get(_LoggerName)
	if !ok {
		return nil
	}

	return logger.(*zap.Logger)
}

func (c *context) setLogger(logger *zap.Logger) {
	c.ctx.Set(_LoggerName, logger)
}

func (c *context) getPayload() errno.Error {
	payload, _ := c.ctx.Get(_PayloadName)
	return payload.(errno.Error)
}

func (c *context) SetPayload(payload errno.Error) {
	c.ctx.Set(_PayloadName, payload)
}

func (c *context) Success(data interface{}) {
	lang := c.Lang()
	succ := errno.NewError(errcode.OK, errcode.ErrorMsg[lang][errcode.OK].GetMsg())
	c.ctx.Set(_PayloadName, succ.WithData(data))
}
func (c *context) Fail(code int) {
	lang := c.Lang()

	if err, ok := errcode.ErrorMsg[lang][code]; ok {
		c.ctx.Set(_PayloadName, err)
	} else if err, ok := errcode.ErrorMsg[errcode.DEFAULT_LANG][code]; ok {
		c.ctx.Set(_PayloadName, err)
	} else {
		c.ctx.Set(_PayloadName, errcode.ErrorMsg[errcode.DEFAULT_LANG][errcode.ERR_UNKNOWN])
	}
}

func (c *context) Header() http.Header {
	header := c.ctx.Request.Header

	clone := make(http.Header, len(header))
	for k, v := range header {
		value := make([]string, len(v))
		copy(value, v)

		clone[k] = value
	}
	return clone
}

func (c *context) GetHeader(key string) string {
	return c.ctx.GetHeader(key)
}

func (c *context) SetHeader(key, value string) {
	c.ctx.Header(key, value)
}

func (c *context) UserID() int {
	val, ok := c.ctx.Get(_UserID)
	if !ok {
		return 0
	}

	return val.(int)
}
func (c *context) Lang() string {
	val, ok := c.ctx.Get(_Lang)
	if !ok {
		return "en"
	}

	return val.(string)
}

func (c *context) setUserID(userID int) {
	c.ctx.Set(_UserID, userID)
}

func (c *context) UserName() string {
	val, ok := c.ctx.Get(_UserName)
	if !ok {
		return ""
	}

	return val.(string)
}

func (c *context) setUserName(userName string) {
	c.ctx.Set(_UserName, userName)
}

func (c *context) AbortWithError(err errno.Error) {
	if err != nil {
		httpCode := err.GetHttpCode()
		if httpCode == 0 {
			httpCode = http.StatusInternalServerError
		}

		c.ctx.AbortWithStatus(httpCode)
		c.ctx.Set(_AbortErrorName, err)
	}
}

func (c *context) abortError() errno.Error {
	err, _ := c.ctx.Get(_AbortErrorName)
	return err.(errno.Error)
}

func (c *context) Alias() string {
	path, ok := c.ctx.Get(_Alias)
	if !ok {
		return ""
	}

	return path.(string)
}

func (c *context) setAlias(path string) {
	if path = strings.TrimSpace(path); path != "" {
		c.ctx.Set(_Alias, path)
	}
}

func (c *context) RawData() []byte {
	body, ok := c.ctx.Get(_BodyName)
	if !ok {
		return nil
	}

	return body.([]byte)
}

// Method 请求的method
func (c *context) Method() string {
	return c.ctx.Request.Method
}

// Host 请求的host
func (c *context) Host() string {
	return c.ctx.Request.Host
}

// Path 请求的路径(不附带querystring)
func (c *context) Path() string {
	return c.ctx.Request.URL.Path
}

// URI unescape后的uri
func (c *context) URI() string {
	uri, _ := url.QueryUnescape(c.ctx.Request.URL.RequestURI())
	return uri
}

func (c *context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *context) Done() <-chan struct{} {
	return nil
}

func (c *context) Err() error {
	return nil
}

func (c *context) Value(key interface{}) interface{} {
	if key == 0 {
		return c.ctx.Request
	}
	if keyAsString, ok := key.(string); ok {
		val, _ := c.ctx.Get(keyAsString)
		return val
	}
	return nil
}
