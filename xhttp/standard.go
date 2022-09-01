package xhttp

import (
	sctx "context"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/container"
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/golibs/httputil"
)

const TypeNode = "xhttp"
const _defaultName = "default"

type StandardHttp interface {
	GetHttp(name ...string) (c Client)
}

type Client interface {
	//Swap 将当前请求参数作为RPC参数并发送RPC请求
	Swap(ctx context.Context, service string, opts ...RequestOption) (res Body, err error)

	//RequestByCtx RPC请求，可通过context撤销请求
	Request(ctx sctx.Context, service string, input interface{}, opts ...RequestOption) (res Body, err error)
}

type Body = httputil.Body

//xHttp 服务
type xHttp struct {
	container container.Container
}

//NewXhttp 服务代理
func NewXhttp(container container.Container) StandardHttp {
	return &xHttp{
		container: container,
	}
}

//GetRPC 获取缓存操作对象
func (s *xHttp) GetHttp(name ...string) (c Client) {
	realName := _defaultName
	if len(name) > 0 {
		realName = name[0]
	}
	if realName == "" {
		realName = _defaultName
	}

	obj, err := s.container.GetOrCreate(TypeNode, realName, func(cfg config.Config) (interface{}, error) {
		dbcfg := cfg.Get(TypeNode).Get(realName)
		return newXhttp(realName, dbcfg)
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
	return NewXhttp(c)
}
