package user_handler

import (
	"gin-api/internal/api/code"
	"gin-api/internal/api/model/user_model"
	"gin-api/internal/api/repository/cache_repo"
	"gin-api/internal/api/repository/db_repo"
	"gin-api/internal/api/service/user_service"
	"gin-api/internal/pkg/config"
	"gin-api/internal/pkg/core"
	"gin-api/pkg/token"

	"go.uber.org/zap"
)

var _ UserDemo = (*userDemo)(nil)

type UserDemo interface {
	// i 为了避免被其他包实现
	i()
	// 创建用户
	Create() core.HandlerFunc
	// 通过用户主键ID更新用户昵称
	UpdateNickNameByID() core.HandlerFunc
	// 用户登录
	Login() core.HandlerFunc
	// 用户详情
	Detail() core.HandlerFunc
}

type userDemo struct {
	logger      *zap.Logger
	userService user_service.UserService
}

func NewUserDemo(logger *zap.Logger, db db_repo.Repo, cache cache_repo.Repo) UserDemo {
	return &userDemo{
		logger:      logger,
		userService: user_service.NewUserService(db, cache),
	}
}

func (u *userDemo) i() {}

// 创建用户
// @Summary 创建用户
// @Description 创建用户
// @Tags Demo
// @Accept  json
// @Produce  json
// @Param RequestInfo body user_model.CreateRequest true "请求信息"
// @Success 200 {object} user_model.CreateResponse "返回信息"
// @Router /user/create [post]
func (u *userDemo) Create() core.HandlerFunc {
	return func(c core.Context) {
		req := new(user_model.CreateRequest)
		res := new(user_model.CreateResponse)
		if err := c.ShouldBindJSON(req); err != nil {
			u.logger.Error("[user] should bind json err", zap.Error(err))
			c.Fail(errcode.ERR_PARAM_BIND)
			return
		}

		id, err := u.userService.Create(c, req)
		if err != nil {
			u.logger.Error("[user] Create err", zap.Error(err))
			c.Fail(errcode.ERR_USER_CREATE)
			return
		}

		res.Id = id
		c.Success(res)
	}
}

// 更新用户名称
// @Summary 更新用户名称
// @Description 更新用户名称
// @Tags Demo
// @Accept  json
// @Produce  json
// @Param RequestInfo body user_model.UpdateNickNameByIDRequest true "请求信息"
// @Success 200 {object} user_model.UpdateNickNameByIDResponse "返回信息"
// @Router /user/update [post]
func (u *userDemo) UpdateNickNameByID() core.HandlerFunc {
	return func(c core.Context) {
		req := new(user_model.UpdateNickNameByIDRequest)
		res := new(user_model.UpdateNickNameByIDResponse)
		if err := c.ShouldBindJSON(req); err != nil {
			u.logger.Error("[user] should bind json err", zap.Error(err))
			c.Fail(errcode.ERR_PARAM_BIND)
			return
		}

		err := u.userService.UpdateNickNameByID(c, req.Id, req.NickName)
		if err != nil {
			u.logger.Error("[user] UpdateNickNameByID err", zap.Error(err))
			c.Fail(errcode.ERR_USER_UPDATE)
			return
		}

		res.Id = req.Id
		c.Success(res)
	}
}

// 用户登录
// @Summary 用户登录
// @Description 用户登录
// @Tags Demo
// @Accept  json
// @Produce  json
// @Param RequestInfo body user_model.LoginRequest true "请求信息"
// @Success 200 {object} user_model.LoginResponse "返回信息"
// @Router /user/login [post]
func (u *userDemo) Login() core.HandlerFunc {
	return func(c core.Context) {
		req := new(user_model.LoginRequest)
		res := new(user_model.LoginResponse)
		if err := c.ShouldBindJSON(req); err != nil {
			u.logger.Error("should bind json err", zap.Error(err))
			c.Fail(errcode.ERR_PARAM_BIND)
			return
		}

		cfg := config.GetConfig()
		tokenString, err := token.New(cfg.GetString("jwt.secret")).Sign(req.UserID, req.UserName)
		if err != nil {
			u.logger.Error("token sign err", zap.Error(err))
			c.Fail(errcode.ERR_AUTHORIZATION)
			return
		}

		claims, err := token.New(cfg.GetString("jwt.secret")).Parse(tokenString)
		if err != nil {
			u.logger.Error("token parse err", zap.Error(err))
			c.Fail(errcode.ERR_AUTHORIZATION)
			return
		}

		res.Authorization = tokenString
		res.ExpireTime = claims.ExpiresAt

		c.Success(res)
	}
}

// 用户详情
// @Summary 用户详情
// @Description 用户详情
// @Tags Demo
// @Accept  json
// @Produce  json
// @Param username path string true "用户名"
// @Success 200 {object} user_model.DetailResponse "返回信息"
// @Router /user/info/{username} [get]
func (u *userDemo) Detail() core.HandlerFunc {
	return func(c core.Context) {
		req := new(user_model.DetailRequest)
		res := new(user_model.DetailResponse)
		if err := c.ShouldBindURI(req); err != nil {
			u.logger.Error("should bind uri err", zap.Error(err))
			c.Fail(errcode.ERR_PARAM_BIND)
			return
		}

		user, err := u.userService.GetUserByUserName(c, req.UserName)
		if err != nil {
			u.logger.Error("[user] GetUserByUserName err", zap.Error(err))

			c.Fail(errcode.ERR_USER_SEARCH)
			return
		}

		res.Id = user.Id
		res.UserName = user.UserName
		res.NickName = user.NickName
		res.Mobile = user.Mobile
		c.Success(res)
	}
}
