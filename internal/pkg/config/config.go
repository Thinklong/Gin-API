package config

import ()

// 初始化配置信息 和 日志文件

var configManager = Manager{}
var serviceManager = Manager{}

const serviceName = "gin-api"

func init() {
	// 配置解析
	err := configManager.Init(serviceName)
	if err != nil {
		panic("init config failed, " + err.Error())
	}

}

func GetConfig() *Config {
	c := configManager.GetIns()
	return c
}
