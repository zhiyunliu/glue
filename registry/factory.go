package registry

import (
	"fmt"
)

//IFactory 注册中心构建器
type Factory interface {
	Name() string
	Create(*Options) (Registrar, error)
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

func GetRegistrar(cfg *Config, opts ...Option) (Registrar, error) {
	factory, ok := factoryMap[cfg.Proto]
	if !ok {
		return nil, fmt.Errorf("不支持的Proto类型[%s]", cfg.Proto)
	}

	opt := &Options{
		cfg: cfg,
	}
	for i := range opts {
		opts[i](opt)
	}

	return factory.Create(opt)
}
