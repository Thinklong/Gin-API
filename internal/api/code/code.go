package errcode

import (
	"gin-api/pkg/errno"
	"net/http"
)

const (
	OK                = 0
	ERR_UNKNOWN       = 10200 //unknown
	ERR_PARAM_BIND    = 10201
	ERR_AUTHORIZATION = 10202

	ERR_USER        = 20101
	ERR_USER_CREATE = 20102
	ERR_USER_UPDATE = 20103
	ERR_USER_SEARCH = 20104
	ERR_USER_OTHER  = 20105

	DEFAULT_LANG = "en"
)

var ErrorMsg = map[string]map[int]errno.Error{
	"zh": {
		OK: errno.NewError(OK, "操作成功"),
		// 服务级错误码
		ERR_UNKNOWN:       errno.NewError(ERR_UNKNOWN, "未知错误"),
		ERR_PARAM_BIND:    errno.NewError(ERR_PARAM_BIND, "参数信息有误"),
		ERR_AUTHORIZATION: errno.NewError(ERR_AUTHORIZATION, "签名信息有误"),
		// 模块级错误码 - 用户模块
		ERR_USER:        errno.NewError(ERR_USER, "非法用户"),
		ERR_USER_CREATE: errno.NewError(ERR_USER_CREATE, "创建用户失败"),
		ERR_USER_UPDATE: errno.NewError(ERR_USER_UPDATE, "更新用户失败"),
		ERR_USER_SEARCH: errno.NewError(ERR_USER_SEARCH, "查询用户失败"),
		ERR_USER_OTHER:  errno.NewError(ERR_USER_OTHER, "调用他方接口失败"),
	},
	"en": {
		OK: errno.NewError(OK, "success"),
		// 服务级错误码
		ERR_UNKNOWN:       errno.NewError(ERR_UNKNOWN, "unknown error"),
		ERR_PARAM_BIND:    errno.NewError(ERR_PARAM_BIND, "param bind"),
		ERR_AUTHORIZATION: errno.NewError(ERR_AUTHORIZATION, "authorization bind"),

		// 模块级错误码 - 用户模块
		ERR_USER:        errno.NewError(ERR_USER, "非法用户"),
		ERR_USER_CREATE: errno.NewError(ERR_USER_CREATE, "创建用户失败"),
		ERR_USER_UPDATE: errno.NewError(ERR_USER_UPDATE, "更新用户失败"),
		ERR_USER_SEARCH: errno.NewError(ERR_USER_SEARCH, "查询用户失败"),
		ERR_USER_OTHER:  errno.NewError(ERR_USER_OTHER, "调用他方接口失败"),
	},
}
var ErrServer = errno.SysemError(http.StatusInternalServerError, 10101, http.StatusText(http.StatusInternalServerError))
var ErrManyRequest = errno.SysemError(http.StatusTooManyRequests, 10102, http.StatusText(http.StatusTooManyRequests))
