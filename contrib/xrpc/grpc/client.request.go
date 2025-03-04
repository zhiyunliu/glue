package grpc

import (
	sctx "context"
	"fmt"
	"net/http"
	"net/url"

	cmap "github.com/orcaman/concurrent-map/v2"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/registry"
	"github.com/zhiyunliu/glue/xrpc"
)

// Request RPC Request
type Request struct {
	requests     cmap.ConcurrentMap[string, any]
	clientConfig *clientConfig
}

// NewRequest 构建请求
func NewRequest(clientCfg *clientConfig) *Request {

	req := &Request{
		clientConfig: clientCfg,
		requests:     cmap.New[any](),
	}
	return req
}

// Swap 将当前请求参数作为RPC参数并发送RPC请求
func (r *Request) Swap(ctx context.Context, service string, opts ...xrpc.RequestOption) (res xrpc.Body, err error) {

	//获取内容
	input := ctx.Request().Body().Bytes()
	//复制请求头
	hd := ctx.Request().Header()

	opts = append(opts, xrpc.WithHeaders(hd.Values()), xrpc.WithMethod(ctx.Request().GetMethod()), xrpc.WithQuery(ctx.Request().Query().String()))

	// 发送请求
	return r.Request(ctx.Context(), service, input, opts...)
}

// RequestByCtx RPC请求，可通过context撤销请求
// service=grpc://servername/path
func (r *Request) Request(ctx sctx.Context, service string, input any, opts ...xrpc.RequestOption) (res xrpc.Body, err error) {
	client, err := r.getClient(service)
	if err != nil {
		return
	}

	nopts := make([]xrpc.RequestOption, 0, len(opts)+2)
	nopts = append(nopts, opts...)
	nopts = append(nopts, xrpc.WithSourceName())

	if logger, ok := log.FromContext(ctx); ok {
		nopts = append(nopts, xrpc.WithXRequestID(logger.SessionID()))
	}

	//处理可选参数
	opt := &xrpc.Options{
		Method: http.MethodPost,
		Header: make(map[string]string),
	}
	for i := range nopts {
		nopts[i](opt)
	}

	if client.setting.Trace {
		ctx, span := client.tracer.Start(ctx, client.reqPath.Path, opt.Header)
		defer func() {
			if err != nil {
				client.tracer.End(ctx, span, err)
				return
			}
			status := int32(http.StatusOK)
			if res != nil {
				status = res.GetStatus()
			}
			client.tracer.End(ctx, span, status)
		}()
	}

	if opt.StreamProcessor == nil {
		return client.RequestByString(ctx, input, opt)
	}
	return client.RequestByStream(ctx, input, opt)
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

func (r *Request) getClient(service string) (client *Client, err error) {
	pathVal, err := url.Parse(service)
	if err != nil {
		err = fmt.Errorf("grpc.Request url.Parse=%s,Error:%w", service, err)
		return
	}

	//todo:当前是通过url 进行client 构建，是否考虑只通过服务来构建客户端？
	key := fmt.Sprintf("%s:%s", r.clientConfig.Name, service)
	cientObj := r.requests.Upsert(key, pathVal, func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
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
		tclient, err := NewClient(registrar, r.clientConfig, pathVal)
		if err != nil {
			panic(err)
		}
		return tclient
	})

	return cientObj.(*Client), nil
}
