package http

import (
	sctx "context"
	"fmt"
	"net/http"
	"net/url"

	cmap "github.com/orcaman/concurrent-map/v2"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/registry"
	"github.com/zhiyunliu/glue/xhttp"
)

// Request RPC Request
type Request struct {
	requests cmap.ConcurrentMap[string, any]
	setting  *setting
}

// NewRequest 构建请求
func NewRequest(setting *setting) *Request {
	req := &Request{
		setting:  setting,
		requests: cmap.New[any](),
	}
	return req
}

// Swap 将当前请求参数作为RPC参数并发送RPC请求
func (r *Request) Swap(ctx context.Context, service string, opts ...xhttp.RequestOption) (res xhttp.Body, err error) {

	//获取内容
	input := ctx.Request().Body().Bytes()
	opts = append(opts, xhttp.WithXRequestID(ctx.Request().GetHeader(constants.HeaderRequestId)))

	//复制请求头
	hd := ctx.Request().Header()

	opts = append(opts, xhttp.WithHeaders(hd.Values()))

	// 发送请求
	return r.Request(ctx.Context(), service, input, opts...)
}

// RequestByCtx RPC请求，可通过context撤销请求
// service=http://servername/path
func (r *Request) Request(ctx sctx.Context, service string, input interface{}, opts ...xhttp.RequestOption) (res xhttp.Body, err error) {

	pathVal, err := url.Parse(service)
	if err != nil {
		err = fmt.Errorf("http.Request url.Parse=%s,Error:%w", service, err)
		return
	}

	client := r.getClient(pathVal)
	nopts := make([]xhttp.RequestOption, 0, len(opts)+2)
	nopts = append(nopts, opts...)
	nopts = append(nopts, xhttp.WithSourceName())

	if logger, ok := log.FromContext(ctx); ok {
		nopts = append(nopts, xhttp.WithXRequestID(logger.SessionID()))
	}

	opt := &xhttp.Options{
		Method: http.MethodGet,
		Header: make(map[string]string),
	}
	for i := range nopts {
		nopts[i](opt)
	}

	return client.RequestByString(ctx, pathVal, input, opt)
}

// Close 关闭RPC连接
func (r *Request) Close() error {
	r.requests.IterCb(func(key string, v any) {
		v.(*Client).Close()
	})
	r.requests.Clear()
	return nil
}

func (r *Request) getClient(pathVal *url.URL) *Client {
	//todo:当前是通过url 进行client 构建，是否考虑只通过服务来构建客户端？
	key := fmt.Sprintf("%s:%s.%s", r.setting.Name, pathVal.Host, pathVal.Scheme)
	tmpClient := r.requests.Upsert(key, pathVal, func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
		if exist {
			return valueInMap
		}
		var (
			registrar registry.Registrar
			err       error
		)
		regCfg := registry.GetRegistrarName(global.Config)
		if regCfg != "" {
			registrar, err = registry.GetRegistrar(global.Config)
			if err != nil {
				panic(err)
			}
		}
		reqPath := newValue.(*url.URL)
		client, err := NewClient(registrar, r.setting, reqPath)
		if err != nil {
			panic(err)
		}
		return client
	})
	return tmpClient.(*Client)
}
