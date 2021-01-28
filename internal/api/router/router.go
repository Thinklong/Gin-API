package router

import (
	"gin-api/internal/api/controller/demo"
	"gin-api/internal/api/controller/user_handler"
	"gin-api/internal/api/repository/cache_repo"
	"gin-api/internal/api/repository/db_repo"
	auth "gin-api/internal/api/router/middleware"
	"gin-api/internal/pkg/core"
	"gin-api/internal/pkg/metrics"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func NewHTTPMux(logger *zap.Logger, db db_repo.Repo, cache cache_repo.Repo) (core.Mux, error) {

	if logger == nil {
		return nil, errors.New("logger required")
	}

	mux, err := core.New(logger,
		core.WithEnableCors(),
		core.WithEnableRate(),
		//core.WithPanicNotify(notify.OnPanicNotify),
		core.WithRecordMetrics(metrics.RecordMetrics),
	)

	if err != nil {
		panic(err)
	}

	demoHandler := demo.NewDemo(logger)
	userHandler := user_handler.NewUserDemo(logger, db, cache)

	u := mux.Group("/user")
	{
		u.POST("/login", userHandler.Login())
		u.POST("/create", userHandler.Create())
		u.GET("/info/:username", core.AliasForRecordMetrics("/user/info"), userHandler.Detail())
		u.POST("/update", userHandler.UpdateNickNameByID())
	}

	d := mux.Group("/demo", core.WrapAuthHandler(auth.AuthHandler)) //使用 auth 验证
	{
		d.GET("user/:name", core.AliasForRecordMetrics("/demo/user"), demoHandler.User())

		// 模拟数据
		d.GET("get/:name", core.AliasForRecordMetrics("/demo/get"), demoHandler.Get())
		d.POST("post", demoHandler.Post())
	}

	return mux, nil
}
