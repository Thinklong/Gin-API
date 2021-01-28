package middleware

import (
	"gin-api/internal/api/code"
	"gin-api/internal/pkg/config"
	"gin-api/internal/pkg/core"
	"gin-api/pkg/errno"
	"gin-api/pkg/token"
)

func AuthHandler(ctx core.Context) (userId int, userName string, err errno.Error) {
	auth := ctx.GetHeader("Authorization")
	lang := ctx.Lang()
	if auth == "" {
		err = errcode.ErrorMsg[lang][errcode.ERR_AUTHORIZATION]
		return
	}

	cfg := config.GetConfig()
	claims, errParse := token.New(cfg.GetString("jwt.secret")).Parse(auth)
	if errParse != nil {
		err = errcode.ErrorMsg[lang][errcode.ERR_AUTHORIZATION]
		return
	}

	userId = claims.UserID
	if userId <= 0 {
		err = errcode.ErrorMsg[lang][errcode.ERR_AUTHORIZATION]
		return
	}
	userName = claims.UserName
	return
}
