package grpc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/balancer"
	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/middleware/tracing"
	"github.com/zhiyunliu/glue/registry"
	"github.com/zhiyunliu/glue/xrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

type Client struct {
	registrar       registry.Registrar
	setting         *clientConfig
	reqPath         *url.URL
	conn            *grpc.ClientConn
	client          grpcproto.GRPCClient
	balancerBuilder resolver.Builder
	ctx             context.Context
	ctxCancel       context.CancelFunc
	tracer          *tracing.Tracer
}

// NewClient 创建RPC客户端,地址是远程RPC服务器地址或注册中心地址
func NewClient(registrar registry.Registrar, setting *clientConfig, reqPath *url.URL) (*Client, error) {
	client := &Client{
		registrar: registrar,
		setting:   setting,
		reqPath:   reqPath,
	}
	client.tracer = tracing.NewTracer(trace.SpanKindClient)

	client.ctx, client.ctxCancel = context.WithCancel(context.Background())

	err := client.connect()
	if err != nil {
		err = fmt.Errorf("grpc.connect失败:%s(%v)(err:%v)", reqPath.String(), client.setting.ConnTimeout, err)
		return nil, err
	}
	return client, nil
}

// RequestByString 发送Request请求
func (c *Client) RequestByString(ctx context.Context, input []byte, opts ...xrpc.RequestOption) (res xrpc.Body, err error) {
	//处理可选参数
	o := &xrpc.Options{
		Method: http.MethodPost,
		Header: make(map[string]string),
	}
	for _, opt := range opts {
		opt(o)
	}
	if c.setting.Trace {
		ctx, span := c.tracer.Start(ctx, c.reqPath.Path, o.Header)
		defer func() {
			if err != nil {
				c.tracer.End(ctx, span, err)
				return
			}
			c.tracer.End(ctx, span, res.GetStatus())
		}()
	}

	response, err := c.clientRequest(ctx, o, input)
	if err != nil {
		return newBodyByError(err), err
	}
	return response, err
}

// RequestByString 发送Request请求
func (c *Client) RequestByStream(ctx context.Context, processor xrpc.StreamProcessor, opts ...xrpc.RequestOption) (err error) {
	//处理可选参数
	o := &xrpc.Options{
		Method: http.MethodPost,
		Header: make(map[string]string),
	}
	for _, opt := range opts {
		opt(o)
	}

	servicePath := c.reqPath.Path
	if len(o.Query) > 0 {
		servicePath = fmt.Sprintf("%s?%s", servicePath, o.Query)
	}
	steamClient, err := c.client.StreamProcess(ctx, grpc.WaitForReady(o.WaitForReady))
	if err != nil {
		return err
	}

	//发送服务分发数据信息
	err = steamClient.Send(&grpcproto.Request{
		Method:  o.Method,
		Service: servicePath,
		Header:  o.Header,
	})
	if err != nil {
		return err
	}
	err = processor(&grpcClientStreamRequest{
		servicePath:  servicePath,
		header:       o.Header,
		method:       o.Method,
		streamClient: steamClient,
	})

	return err
}

// Close 关闭RPC客户端连接
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.ctxCancel()
	}
}

// Connect 连接到RPC服务器，如果当前无法连接系统会定时自动重连
// 未使用压缩，由于传输数据默认限制为4M(已修改为20M)压缩后会影响系统并发能力
// grpc.WithDefaultCallOptions(grpc.UseCompressor(Snappy)),
// grpc.WithDecompressor(grpc.NewGZIPDecompressor()),
// grpc.WithCompressor(grpc.NewGZIPCompressor()),
func (c *Client) connect() (err error) {
	c.balancerBuilder = balancer.NewRegistrarBuilder(c.ctx, c.registrar, c.reqPath)

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(c.setting.ConnTimeout)*time.Second)
	c.conn, err = grpc.DialContext(ctx,
		c.reqPath.String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(string(c.setting.ServerConfig)),
		//grpc.WithBalancerName(c.setting.Balancer),
		grpc.WithResolvers(c.balancerBuilder),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(Snappy)),
	)

	if err != nil {
		return fmt.Errorf("grpc.DialContext:path=%s.Error:%s", c.reqPath.String(), err)
	}
	c.client = grpcproto.NewGRPCClient(c.conn)
	return nil
}

func (c *Client) clientRequest(ctx context.Context, o *xrpc.Options, input []byte) (response *grpcproto.Response, err error) {
	servicePath := c.reqPath.Path
	if len(o.Query) > 0 {
		servicePath = fmt.Sprintf("%s?%s", servicePath, o.Query)
	}
	return c.client.Process(ctx,
		&grpcproto.Request{
			Method:  o.Method, //借用http的method
			Service: servicePath,
			Header:  o.Header,
			Body:    input,
		},
		grpc.WaitForReady(o.WaitForReady))

}
