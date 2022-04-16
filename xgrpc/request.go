package xgrpc

import (
	sctx "context"
	"encoding/json"
	"fmt"
	"net/url"

	cmap "github.com/orcaman/concurrent-map"

	"github.com/zhiyunliu/gel/constants"
	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/gel/global"
	"github.com/zhiyunliu/gel/registry"
	"github.com/zhiyunliu/golibs/bytesconv"
)

//IRequest Component rpc
type IRequest interface {

	//Swap 将当前请求参数作为RPC参数并发送RPC请求
	Swap(ctx context.Context, service string, opts ...RequestOption) (res Body, err error)

	//RequestByCtx RPC请求，可通过context撤销请求
	Request(ctx sctx.Context, service string, input interface{}, opts ...RequestOption) (res Body, err error)
}

//Request RPC Request
type Request struct {
	requests cmap.ConcurrentMap
	setting  *setting
}

//NewRequest 构建请求
func NewRequest(setting *setting) *Request {
	req := &Request{
		setting:  setting,
		requests: cmap.New(),
	}
	return req
}

//Swap 将当前请求参数作为RPC参数并发送RPC请求
func (r *Request) Swap(ctx context.Context, service string, opts ...RequestOption) (res Body, err error) {

	//获取内容
	input := ctx.Request().Body().Bytes()
	opts = append(opts, WithTraceID(ctx.Request().GetHeader(constants.HeaderRequestId)))

	//复制请求头
	hd := ctx.Request().Header()

	opts = append(opts, WithHeaders(hd))

	// 发送请求
	return r.Request(ctx.Context(), service, input, opts...)
}

//RequestByCtx RPC请求，可通过context撤销请求
//service=grpc://servername/path
func (r *Request) Request(ctx sctx.Context, service string, input interface{}, opts ...RequestOption) (res Body, err error) {

	pathVal, err := url.Parse(service)
	if err != nil {
		err = fmt.Errorf("grpc.Request url.Parse=%s,Error:%w", service, err)
		return
	}

	//如果入参不是ip 通过注册中心去获取所请求平台的所有rpc服务子节点  再通过路由匹配获取真实的路由
	key := fmt.Sprintf("%s:%s", r.setting.Name, service)
	tmpClient := r.requests.Upsert(key, pathVal, func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
		if exist {
			return valueInMap
		}

		registrar, err := registry.GetRegistrar(global.Config)
		if err != nil {
			panic(err)
		}

		pathVal = newValue.(*url.URL)
		client, err := NewClient(registrar, r.setting, pathVal)
		if err != nil {
			panic(err)
		}
		return client
	})

	client := tmpClient.(*Client)
	nopts := make([]RequestOption, 0, len(opts)+1)
	nopts = append(nopts, opts...)
	if reqidVal := ctx.Value(constants.HeaderRequestId); reqidVal != nil {
		nopts = append(nopts, WithTraceID(fmt.Sprintf("%+v", reqidVal)))
	}

	var bodyBytes []byte

	switch t := input.(type) {
	case []byte:
		bodyBytes = t
	case string:
		bodyBytes = bytesconv.StringToBytes(t)
	case *string:
		bodyBytes = bytesconv.StringToBytes(*t)
	default:
		bodyBytes, _ = json.Marshal(t)
		nopts = append(nopts, WithContentType(constants.ContentTypeApplicationJSON))
	}

	return client.RequestByString(ctx, bodyBytes, nopts...)
}

//Close 关闭RPC连接
func (r *Request) Close() error {
	r.requests.IterCb(func(key string, v interface{}) {
		client := v.(*Client)
		client.Close()
	})
	r.requests.Clear()
	return nil
}
