package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/balancer"
	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/registry"
	"github.com/zhiyunliu/glue/xrpc"
	"github.com/zhiyunliu/golibs/bytesconv"
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
}

// NewClient 创建RPC客户端,地址是远程RPC服务器地址或注册中心地址
func NewClient(registrar registry.Registrar, setting *clientConfig, reqPath *url.URL) (*Client, error) {
	client := &Client{
		registrar: registrar,
		setting:   setting,
		reqPath:   reqPath,
	}

	client.ctx, client.ctxCancel = context.WithCancel(context.Background())

	err := client.connect()
	if err != nil {
		err = fmt.Errorf("grpc.connect失败:%s(%v)(err:%v)", reqPath.String(), client.setting.ConnTimeout, err)
		return nil, err
	}
	return client, nil
}

// RequestByString 发送Request请求
func (c *Client) RequestByString(ctx context.Context, input any, opts *xrpc.Options) (res xrpc.Body, err error) {

	var bodyBytes []byte
	switch t := input.(type) {
	case []byte:
		bodyBytes = t
	case string:
		bodyBytes = bytesconv.StringToBytes(t)
	case *string:
		bodyBytes = bytesconv.StringToBytes(*t)
	default:
		bodyBytes, _ = json.Marshal(input)
		xrpc.WithContentType(constants.ContentTypeApplicationJSON)(opts)
	}

	response, err := c.clientRequest(ctx, opts, bodyBytes)
	if err != nil {
		return newBodyByError(err), err
	}
	return response, err
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

	c.conn, err = grpc.NewClient(
		c.reqPath.String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(string(c.setting.ServerConfig)),
		grpc.WithResolvers(c.balancerBuilder),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(Snappy)),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: time.Duration(c.setting.ConnTimeout) * time.Second,
		}),
	)

	if err != nil {
		return fmt.Errorf("grpc.DialContext:path=%s.Error:%s", c.reqPath.String(), err)
	}
	c.client = grpcproto.NewGRPCClient(c.conn)
	return nil
}

func (c *Client) buildGrpcOpts(opts *xrpc.Options) []grpc.CallOption {
	grpcOpts := []grpc.CallOption{
		grpc.WaitForReady(opts.WaitForReady),
	}
	if opts.MaxCallRecvMsgSize > 0 {
		grpcOpts = append(grpcOpts, grpc.MaxCallRecvMsgSize(opts.MaxCallRecvMsgSize))
	}
	if opts.MaxCallSendMsgSize > 0 {
		grpcOpts = append(grpcOpts, grpc.MaxCallSendMsgSize(opts.MaxCallSendMsgSize))
	}
	return grpcOpts
}

func (c *Client) clientRequest(ctx context.Context, o *xrpc.Options, bodyBytes []byte) (response *grpcproto.Response, err error) {
	servicePath := c.reqPath.Path
	if len(o.Query) > 0 {
		servicePath = fmt.Sprintf("%s?%s", servicePath, o.Query)
	}

	return c.client.Process(ctx,
		&grpcproto.Request{
			Method:  o.Method, //借用http的method
			Service: servicePath,
			Header:  o.Header,
			Body:    bodyBytes,
		},
		c.buildGrpcOpts(o)...)

}

// RequestByString 发送Request请求
func (c *Client) RequestByStream(ctx context.Context, input any, opts *xrpc.Options) (body xrpc.Body, err error) {

	switch processor := opts.StreamProcessor.(type) {
	//双向流
	case xrpc.BidirectionalStreamProcessor:
		return xrpc.NewEmptyBody(), c.BidirectionalStreamProcessor(ctx, processor, opts)
	case func(xrpc.BidirectionalStreamClient) error:
		return xrpc.NewEmptyBody(), c.BidirectionalStreamProcessor(ctx, processor, opts)

	//客户端流
	case func(xrpc.ClientStreamClient) (err error):
		return c.ClientStreamProcessor(ctx, processor, opts)
	case xrpc.ClientStreamProcessor:
		return c.ClientStreamProcessor(ctx, processor, opts)

	//服务端流
	case func(xrpc.ServerStreamClient) (err error):
		return xrpc.NewEmptyBody(), c.ServerStreamProcessor(ctx, processor, input, opts)
	case xrpc.ServerStreamProcessor:
		return xrpc.NewEmptyBody(), c.ServerStreamProcessor(ctx, processor, input, opts)

	//默认流处理器--使用客户端流
	case xrpc.DefaultProcessor:
		defaultProcessor, err := xrpc.BuildDefaultClientStreamProcess(input)
		if err != nil {
			return xrpc.NewEmptyBody(), err
		}
		return c.ClientStreamProcessor(ctx, defaultProcessor, opts)
	default:

	}
	return xrpc.NewEmptyBody(), fmt.Errorf("未知的xrpc.StreamProcessor类型")
}
