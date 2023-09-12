package config

import (
	"fmt"
	"strings"

	"github.com/zhiyunliu/golibs/xnet"
)

// IFactory 注册中心构建器
type Factory interface {
	Name() string
	Create(Config) (Source, error)
}

var factoryMap = map[string]Factory{}

// Register 添加注册中心工厂对象
func Register(factory Factory) {
	if factory == nil {
		panic(fmt.Errorf("registry: factory is nil"))
	}
	name := factory.Name()
	if _, ok := factoryMap[name]; ok {
		panic(fmt.Errorf("registry: factory called twice:" + name))
	}
	factoryMap[name] = factory
}

func GetConfigName(cfg Config) string {
	return cfg.Value("config").String()
}

func GetConfig(cfg Config) (Source, error) {
	cfgVal := GetConfigName(cfg)
	if strings.TrimSpace(cfgVal) == "" {
		return nil, nil
	}

	//nacos://xxxx
	protoName, configName, err := xnet.Parse(cfgVal)
	if err != nil {
		return nil, err
	}
	factory, ok := factoryMap[protoName]
	if !ok {
		return nil, fmt.Errorf("config 不支持的Proto类型[%s]", protoName)
	}
	regCfg := cfg.Get(protoName).Get(configName)

	return factory.Create(regCfg)
}
