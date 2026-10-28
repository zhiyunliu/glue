package xipc

import (
	"context"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
)

// ServerResover 定义配置文件转换方法
type ServerResolver interface {
	Name() string
	Resolve(name string,
		router *engine.RouterGroup,
		cfg config.Config,
		opts ...engine.Option) (Server, error)
}


// Server 是 IPC 传输实现的生命周期契约。
type Server interface {
	GetProto() string
	GetAddr() string
	Serve(context.Context) error
	Stop(context.Context) error
}