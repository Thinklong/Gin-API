package config

import (
	"errors"
	"fmt"
	"gin-api/internal/pkg/consul"
	"os"
	"time"
)

type Manager struct {
	name    string
	config  *Config
	success bool
}

func (manager *Manager) Init(serviceName string) error {
	manager.name = serviceName
	env := os.Getenv("APP_ENV")
	var err error
	if env == "PROD" {
		// 使用consul进行服务注册和配置管理
		if err = manager.WatchConfig(); err != nil {
			return err
		}
	} else {
		// 使用本地配置文件
		if manager.config, err = GetConfigFromLocal(); err != nil {
			return err
		}
	}
	return nil
}

func (manager *Manager) GetIns() *Config {
	return manager.config
}

func (manager *Manager) WatchConfig() error {
	var serviceManager = consul.Manager{}
	st := time.Now().UnixNano() / 1e6
	// 配置管理
	go func() {
		configKey := os.Getenv("APP_CONSUL_CONFIG_KEY")
		if configKey == "" {
			configKey = manager.name
		}
		err := serviceManager.WatchKey(configKey, manager.ValueChangeCallback)
		if err != nil {
			// if err, finally got timeout error
			println("config: watch consul env key filed %v" + err.Error())
		}
	}()
	for !manager.success {
		time.Sleep(time.Millisecond * 10)
		if time.Now().UnixNano()/1e6-1500 > st {
			return errors.New("get config from consul timeout, 1500 ms")
		}
	}
	return nil
}

func (manager *Manager) ValueChangeCallback(key string, value []byte) {
	fmt.Println("go-base-config: receive new config data from consul, key " + key)
	manager.config = GetConfigFromContent(value)
	manager.success = true
}
