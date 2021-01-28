package demo

import (
	"time"

	"gin-api/internal/api/code"
	"gin-api/internal/api/repository/third_party_request/go_gin_api_repo"
	"gin-api/internal/pkg/core"
	"gin-api/pkg/httpclient"

	"go.uber.org/zap"
)

type Demo struct {
	logger *zap.Logger
}

func NewDemo(logger *zap.Logger) *Demo {
	return &Demo{
		logger: logger,
	}
}

func (d *Demo) Get() core.HandlerFunc {
	type request struct {
		Name string `uri:"name"`
	}

	type response struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name"`
		Job  string `json:"job"`
	}

	return func(c core.Context) {
		req := new(request)
		if err := c.ShouldBindURI(req); err != nil {
			c.Fail(errcode.ERR_PARAM_BIND)
			return
		}

		if req.Name != "Tom" {
			c.Fail(errcode.ERR_USER)
			return
		}

		c.Success(&response{
			Name: "Tom",
			Job:  "Student",
		})
	}
}

func (d *Demo) Post() core.HandlerFunc {
	type request struct {
		Name string `form:"name"`
	}

	type response struct {
		Name string `json:"name"`
		Job  string `json:"job"`
	}

	return func(c core.Context) {
		req := new(request)
		if err := c.ShouldBindPostForm(req); err != nil {
			c.Fail(errcode.ERR_PARAM_BIND)
			return
		}

		if req.Name != "Jack" {
			c.Fail(errcode.ERR_USER)
			return
		}

		c.Success(&response{
			Name: "Jack",
			Job:  "Teacher",
		})
	}
}

type request struct {
	Name string `uri:"name"`
}

type response []struct {
	Name string `json:"name"` //用户名
	Job  string `json:"job"`  //工作
}

// Get 获取用户信息
// @Summary 获取用户信息
// @Description 获取用户信息
// @Tags Demo
// @Accept  json
// @Produce  json
// @Param name path string true "用户名(Tom)"
// @Param Authorization header string true "签名"
// @Success 200 {object} response "用户信息"
// @Router /demo/user/{name} [get]
func (d *Demo) User() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request)
		if err := c.ShouldBindURI(req); err != nil {
			c.Fail(errcode.ERR_PARAM_BIND)
			return
		}

		if req.Name != "Tom" {
			c.Fail(errcode.ERR_USER)
			return
		}

		res1, err := go_gin_api_repo.DemoGet(req.Name,
			httpclient.WithTTL(time.Second*5),
			httpclient.WithTrace(c.Trace()),
			httpclient.WithLogger(c.Logger()),
			httpclient.WithHeader("Authorization", c.GetHeader("Authorization")),
			httpclient.WithOnFailedRetry(3, time.Second*1, go_gin_api_repo.DemoGetRetryVerify),
		)

		if err != nil {
			d.logger.Error("get [demo/get] err", zap.Error(err))
			c.Fail(errcode.ERR_USER_OTHER)
			return
		}

		res2, err := go_gin_api_repo.DemoPost("Jack",
			httpclient.WithTTL(time.Second*5),
			httpclient.WithTrace(c.Trace()),
			httpclient.WithLogger(c.Logger()),
			httpclient.WithHeader("Authorization", c.GetHeader("Authorization")),
			httpclient.WithOnFailedRetry(3, time.Second*1, go_gin_api_repo.DemoPostRetryVerify),
		)

		if err != nil {
			d.logger.Error("post [demo/post] err", zap.Error(err))
			c.Fail(errcode.ERR_USER_OTHER)
			return
		}

		data := &response{
			{
				Name: res1.Data.Name,
				Job:  res1.Data.Job,
			},
			{
				Name: res2.Data.Name,
				Job:  res2.Data.Job,
			},
		}
		c.Success(data)
	}
}
