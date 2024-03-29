package grpc

import (
	sctx "context"
	"encoding/json"
	"fmt"
	"net/url"

	cmap "github.com/orcaman/concurrent-map"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/registry"
	"github.com/zhiyunliu/glue/xrpc"
	"github.com/zhiyunliu/golibs/bytesconv"
)

// Request RPC Request
type Request struct {
	requests cmap.ConcurrentMap
	setting  *setting
}

// NewRequest 构建请求
func NewRequest(setting *setting) *Request {

	req := &Request{
		setting:  setting,
		requests: cmap.New(),
	}
	return req
}

// Swap 将当前请求参数作为RPC参数并发送RPC请求
func (r *Request) Swap(ctx context.Context, service string, opts ...xrpc.RequestOption) (res xrpc.Body, err error) {

	//获取内容
	input := ctx.Request().Body().Bytes()
	opts = append(opts, xrpc.WithXRequestID(ctx.Request().GetHeader(constants.HeaderRequestId)))

	//复制请求头
	hd := ctx.Request().Header()

	opts = append(opts, xrpc.WithHeaders(hd.Values()))

	// 发送请求
	return r.Request(ctx.Context(), service, input, opts...)
}

// RequestByCtx RPC请求，可通过context撤销请求
// service=grpc://servername/path
func (r *Request) Request(ctx sctx.Context, service string, input interface{}, opts ...xrpc.RequestOption) (res xrpc.Body, err error) {

	pathVal, err := url.Parse(service)
	if err != nil {
		err = fmt.Errorf("grpc.Request url.Parse=%s,Error:%w", service, err)
		return
	}

	//todo:当前是通过url 进行client 构建，是否考虑只通过服务来构建客户端？
	key := fmt.Sprintf("%s:%s", r.setting.Name, service)
	tmpClient := r.requests.Upsert(key, pathVal, func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
		if exist {
			return valueInMap
		}

		var registrar registry.Registrar
		var err error
		_, _, ok := xrpc.IsIpPortAddr(pathVal.Host)
		if !ok {
			registrar, err = registry.GetRegistrar(global.Config)
			if err != nil {
				panic(err)
			}
		}

		pathVal = newValue.(*url.URL)
		client, err := NewClient(registrar, r.setting, pathVal)
		if err != nil {
			panic(err)
		}
		return client
	})

	client := tmpClient.(*Client)
	nopts := make([]xrpc.RequestOption, 0, len(opts)+2)
	nopts = append(nopts, xrpc.WithSourceName(global.AppName))
	nopts = append(nopts, opts...)

	if logger, ok := log.FromContext(ctx); ok {
		nopts = append(nopts, xrpc.WithXRequestID(logger.SessionID()))
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
		nopts = append(nopts, xrpc.WithContentType(constants.ContentTypeApplicationJSON))
	}

	return client.RequestByString(ctx, bodyBytes, nopts...)
}

// Close 关闭RPC连接
func (r *Request) Close() error {
	r.requests.IterCb(func(key string, v interface{}) {
		client := v.(*Client)
		client.Close()
	})
	r.requests.Clear()
	return nil
}
