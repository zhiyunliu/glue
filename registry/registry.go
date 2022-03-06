package registry

import (
	"fmt"

	"github.com/zhiyunliu/velocity/server"
)

//IRegistry 注册中心接口
type IRegistry interface {
	Register(server.Runnable)
	Deregister(server.Runnable)
	GetImpl() interface{}
}

//IFactory 注册中心构建器
type IFactory interface {
	Name() string
	Create(*Options) (IRegistry, error)
}

var factoryMap = map[string]IFactory{}

//Register 添加注册中心工厂对象
func Register(factory IFactory) {
	if factory == nil {
		panic(fmt.Errorf("registry: factory is nil"))
	}
	name := factory.Name()
	if _, ok := factoryMap[name]; ok {
		panic(fmt.Errorf("registry: factory called twice:" + name))
	}
	factoryMap[name] = factory
}

func GetRegistry(cfg *Config, opts ...Option) (IRegistry, error) {
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
