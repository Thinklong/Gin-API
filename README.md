```
本框架是基于Gin框架搭建的Golang API框架，并借鉴了一些其他优秀框架的优点
```

### 特点
- 配置文件解析使用Viper库并区分环境，生产环境使用consul
- 使用zap库日志记录，并使用阿里云日志收集
- 错误码统一定义
- 充分利用Golang接口来约束controller层、service层等
- 基于JWT 身份认证
- DB组件使用GORM V2 + go-redis等
- utils封装多种工具和签名、加密等

### 目录

```
├── main.go                      # 入口
├── configs                      # 配置文件
├── init                         # 项目初始脚本
├── internal                     # 业务目录
│   ├── api                      # 业务代码
│   ├── pkg                      # 内部package
├── logs                         # 日志目录
└── pkg                          # 外部 package
```


#### 要求

- Go version >= 1.15
- Global environment configure (Linux/Mac)

```
export GO111MODULE=on
export GOPROXY=https://goproxy.io
```

#### 环境选项

```
-env fat

// dev 开发环境
// fat 测试环境[默认]
// uat 预发布环境
// pro 正式环境
```



