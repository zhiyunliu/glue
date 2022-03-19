package registry

import (
	"fmt"

	"github.com/zhiyunliu/velocity/config"
)

//IFactory 注册中心构建器
type Factory interface {
	Name() string
	Create(config.Config) (Registrar, error)
}

var factoryMap = map[string]Factory{}

//Register 添加注册中心工厂对象
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

func GetRegistrar(cfg config.Config) (Registrar, error) {
	proto := cfg.Value("proto").String()
	factory, ok := factoryMap[proto]
	if !ok {
		return nil, fmt.Errorf("不支持的Proto类型[%s]", proto)
	}
	return factory.Create(cfg)
}
