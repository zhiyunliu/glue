package xgrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/zhiyunliu/gel/registry"
	"github.com/zhiyunliu/gel/xgrpc/balancer"
	"github.com/zhiyunliu/gel/xgrpc/grpcproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

type Client struct {
	registrar       registry.Registrar
	setting         *setting
	reqPath         *url.URL
	conn            *grpc.ClientConn
	client          grpcproto.GRPCClient
	balancerBuilder resolver.Builder
	ctx             context.Context
	ctxCancel       context.CancelFunc
}

//NewClientByConf 创建RPC客户端,地址是远程RPC服务器地址或注册中心地址
func NewClient(registrar registry.Registrar, setting *setting, reqPath *url.URL) (*Client, error) {
	client := &Client{
		registrar: registrar,
		setting:   setting,
		reqPath:   reqPath,
	}
	client.ctx, client.ctxCancel = context.WithCancel(context.Background())

	err := client.connect()
	if err != nil {
		err = fmt.Errorf("grpc.connect失败:%s(%v)(err:%v)", reqPath.String(), client.setting.ConntTimeout, err)
		return nil, err
	}
	time.Sleep(time.Second)
	return client, nil
}

//Request 发送Request请求
func (c *Client) Request(ctx context.Context, input interface{}, opts ...RequestOption) (res Body, err error) {
	//处理可选参数
	buff, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	return c.RequestByString(ctx, buff, opts...)
}

//RequestByString 发送Request请求
func (c *Client) RequestByString(ctx context.Context, input []byte, opts ...RequestOption) (res Body, err error) {
	//处理可选参数
	o := &requestOptions{}
	for _, opt := range opts {
		opt(o)
	}

	response, err := c.clientRequest(ctx, o, input)
	if err != nil {
		return newBodyByError(err), err
	}
	return response, err
}

//Close 关闭RPC客户端连接
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.ctxCancel()
	}
}

//Connect 连接到RPC服务器，如果当前无法连接系统会定时自动重连
//未使用压缩，由于传输数据默认限制为4M(已修改为20M)压缩后会影响系统并发能力
// grpc.WithDefaultCallOptions(grpc.UseCompressor(Snappy)),
// grpc.WithDecompressor(grpc.NewGZIPDecompressor()),
// grpc.WithCompressor(grpc.NewGZIPCompressor()),
func (c *Client) connect() (err error) {
	c.balancerBuilder = balancer.NewRegistrarBuilder(c.ctx, c.registrar, c.reqPath)

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(c.setting.ConntTimeout)*time.Second)
	c.conn, err = grpc.DialContext(ctx,
		c.reqPath.String(),
		grpc.WithInsecure(),
		grpc.WithBalancerName(c.setting.Balancer),
		grpc.WithResolvers(c.balancerBuilder),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(Snappy)),
	)

	if err != nil {
		return fmt.Errorf("grpc.DialContext:path=%s.Error:%s", c.reqPath.String(), err)
	}
	c.client = grpcproto.NewGRPCClient(c.conn)
	return nil
}

func (c *Client) clientRequest(ctx context.Context, o *requestOptions, input []byte) (response *grpcproto.Response, err error) {

	return c.client.Process(ctx,
		&grpcproto.Request{
			Method:  http.MethodPost, //借用http的method
			Service: c.reqPath.Path,
			Header:  o.Header,
			Body:    input,
		},
		grpc.WaitForReady(o.WaitForReady))

}
