package main

import (
	"context"
	"gin-api/internal/api/repository/cache_repo"
	"gin-api/internal/pkg/config"
	"net/http"
	"time"
	
	"gin-api/internal/api/repository/db_repo"
	"gin-api/internal/api/router"
	"gin-api/pkg/logger"
	"gin-api/pkg/shutdown"
	
	"go.uber.org/zap"
)

// @title gin-api docs api
// @version
// @description

// @contact.name
// @contact.url
// @contact.email

// @host 127.0.0.1:9999
// @BasePath
func main() {
	// 初始化日志
	//loggers, err := logger.NewJSONLogger(
	//	logger.WithField("domain", config.GetConfig().GetString("project.name")),
	//	logger.WithTimeLayout("2006-01-02 15:04:05"),
	//	logger.WithFileP(fmt.Sprintf("./logs/%s-access.log", config.GetConfig().GetString("project.name"))),
	//) 
	
	loggers, err := logger.NewJSONLogger(
		logger.WithField("service", config.GetConfig().GetString("project.name")),
		logger.WithField("module", "main"),
	)
	if err != nil {
		panic(err)
	}
	defer loggers.Sync()
	
	// 初始化数据库
	dbRepo, err := db_repo.New()
	if err != nil {
		loggers.Fatal("new db err", zap.Error(err))
	}
	
	// 初始化缓存
	cacheRepo, err := cache_repo.New()
	if err != nil {
		loggers.Fatal("new cache err", zap.Error(err))
	}
	
	// 初始化 HTTP 服务
	mux, err := router.NewHTTPMux(loggers, dbRepo, cacheRepo)
	if err != nil {
		panic(err)
	}
	
	server := &http.Server{
		Addr:    config.GetConfig().GetString("listen.port"),
		Handler: mux,
	}
	
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loggers.Fatal("http server startup err", zap.Error(err))
		}
	}()
	
	// 优雅关闭
	shutdown.NewHook().Close(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		
		if err := server.Shutdown(ctx); err != nil {
			loggers.Fatal("shutdown err", zap.Error(err))
		} else {
			loggers.Info("shutdown success")
		}
	})
}
