package xrpc

import (
	sctx "context"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
	"github.com/zhiyunliu/glue/context"
)

const TypeNode = "rpcs"
const _defaultName = "default"

type StandardRPC interface {
	GetRPC(name ...string) (c Client)
}

type Client interface {
	//Swap 将当前请求参数作为RPC参数并发送RPC请求
	Swap(ctx context.Context, service string, opts ...RequestOption) (res Body, err error)

	//RequestByCtx RPC请求，可通过context撤销请求
	Request(ctx sctx.Context, service string, input interface{}, opts ...RequestOption) (res Body, err error)
}

type Body interface {
	GetStatus() int32
	GetHeader() map[string]string
	GetResult() []byte
}

//StandardRPC rpc服务
type xRPC struct {
	container container.Container
}

//NewStandardRPC 创建RPC服务代理
func NewXRPC(container container.Container) StandardRPC {
	return &xRPC{
		container: container,
	}
}

//GetRPC 获取缓存操作对象
func (s *xRPC) GetRPC(name ...string) (c Client) {
	realName := _defaultName
	if len(name) > 0 {
		realName = name[0]
	}
	if realName == "" {
		realName = _defaultName
	}
	obj, err := s.container.GetOrCreate(TypeNode, realName, func(cfg config.Config) (interface{}, error) {
		dbcfg := cfg.Get(TypeNode).Get(realName)
		return newXRPC(realName, dbcfg)
	})
	if err != nil {
		panic(err)
	}
	return obj.(Client)
}

type xBuilder struct{}

func NewBuilder() container.StandardBuilder {
	return &xBuilder{}
}

func (xBuilder) Name() string {
	return TypeNode
}

func (xBuilder) Build(c container.Container) interface{} {
	return NewXRPC(c)
}
