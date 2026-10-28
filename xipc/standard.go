package xipc

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
)

const (
	// TypeNode 是客户端配置节点。
	TypeNode    = "ipcs"
	defaultName = "default"
)

// StandardIPC 按名称和选项获取缓存客户端。
type StandardIPC interface {
	GetIPC(name string) Client
}

type standardIPC struct{ c container.Container }

// NewStandardIPC 创建标准客户端提供程序。
func NewStandardIPC(c container.Container) StandardIPC {
	return &standardIPC{c: c}
}

func (s *standardIPC) GetIPC(name string) Client {
	if name == "" {
		name = defaultName
	}
	obj, err := s.c.GetOrCreate(TypeNode, name, func(cfg config.Config) (interface{}, error) {
		var clientConfig Config

		ipcCfg := cfg.Get(TypeNode).Get(name)
		if err := ipcCfg.ScanTo(&clientConfig); err != nil {
			return nil, fmt.Errorf("ipc: 配置错误:%s", err)
		}
		if clientConfig.Proto == "" {
			clientConfig.Proto = defaultName
		}
		return NewClient(clientConfig.Proto, ipcCfg)
	})
	if err != nil {
		panic(err)
	}
	return obj.(Client)
}

type builder struct{}

// NewBuilder 创建供根包注册的客户端构建器。
func NewBuilder() container.StandardBuilder             { return builder{} }
func (builder) Name() string                            { return TypeNode }
func (builder) Build(c container.Container) interface{} { return NewStandardIPC(c) }
