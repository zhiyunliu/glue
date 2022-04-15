package xgrpc

import (
	"github.com/zhiyunliu/gel/config"
	"github.com/zhiyunliu/gel/container"
)

const typeNode = "rpcs"
const _defaultName = "default"

//StandardRPC rpc服务
type StandardRPC struct {
	container container.Container
}

//NewStandardRPC 创建RPC服务代理
func NewStandardRPC(container container.Container) *StandardRPC {
	return &StandardRPC{
		container: container,
	}
}

//GetRPC 获取缓存操作对象
func (s *StandardRPC) GetRPC(name ...string) (c IRequest) {
	realName := _defaultName
	if len(name) > 0 {
		realName = name[0]
	}

	obj, err := s.container.GetOrCreate(typeNode, realName, func(cfg config.Config) (interface{}, error) {
		cfgVal := cfg.Get(typeNode).Value(realName)
		setting := &setting{
			ConntTimeout: 10,
			Balancer:     RoundRobin,
		}
		err := cfgVal.Scan(setting)
		if err != nil && err != config.ErrNotFound {
			return nil, err
		}
		setting.Name = realName
		return NewRequest(setting), nil
	})
	if err != nil {
		panic(err)
	}
	return obj.(IRequest)
}
