package grpc

import (
	sctx "context"
	"fmt"
	"net/url"

	"github.com/zhiyunliu/alloter"
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/xrpc"
	"google.golang.org/grpc/peer"
)

var _ alloter.IRequest = (*clientStreamRequest)(nil)
var _ xrpc.ClientStreamRequest = (*clientStreamRequest)(nil)

// Request 处理任务请求
type clientStreamRequest struct {
	ctx      sctx.Context
	firstReq *grpcproto.Request
	url      *url.URL
	method   string
	params   map[string]string
	header   map[string]string
	stream   grpcproto.GRPC_ClientStreamProcessServer
}

// newStreamRequest 构建任务请求
func newClientStreamRequest(stream grpcproto.GRPC_ClientStreamProcessServer) (r *clientStreamRequest, err error) {

	// 接收第一个请求(服务分发数据信息)
	firstReq, err := stream.Recv()
	if err != nil {
		err = fmt.Errorf("server.grpc.stream.Recv(first):%v", err)
		return
	}

	r = &clientStreamRequest{
		firstReq: firstReq,
		stream:   stream,
		method:   firstReq.Method,
		header:   firstReq.Header,
		params:   make(map[string]string),
	}
	if r.header == nil {
		r.header = map[string]string{}
	}
	r.ctx = stream.Context()
	return r, nil
}

// GetName 获取任务名称
func (m *clientStreamRequest) GetName() string {
	return m.firstReq.Service
}

// GetService 服务名()
func (m *clientStreamRequest) GetService() string {
	return m.firstReq.Service
}

// GetService 服务名()
func (m *clientStreamRequest) GetURL() *url.URL {
	if m.url == nil {
		m.url, _ = url.Parse(m.firstReq.Service)
	}
	return m.url
}

// GetMethod 方法名
func (m *clientStreamRequest) GetMethod() string {
	return m.method
}

func (m *clientStreamRequest) Params() map[string]string {
	return m.params
}

func (m *clientStreamRequest) GetHeader() map[string]string {
	return m.header
}

func (m *clientStreamRequest) Body() []byte {
	return []byte{}
}

func (m *clientStreamRequest) GetRemoteAddr() string {
	if p, ok := peer.FromContext(m.ctx); ok {
		return p.Addr.String()
	}
	return m.header[constants.HeaderRemoteHeader]
}

func (m *clientStreamRequest) Context() sctx.Context {
	return m.ctx
}
func (m *clientStreamRequest) WithContext(ctx sctx.Context) {
	m.ctx = ctx
}

// Recv 反序列化请求体
func (m *clientStreamRequest) Recv(obj any, opts ...xrpc.StreamRevcOption) (closed bool, err error) {
	req, err := m.stream.Recv()
	if err != nil {
		if err.Error() != "EOF" {
			err = fmt.Errorf("server.grpc.stream.Recv:%v", err)
			return false, err
		}
		return true, nil
	}

	opt := xrpc.StreamRecvOptions{
		Unmarshal: defaultUnmarshaler,
	}
	for _, o := range opts {
		o(&opt)
	}
	return false, opt.Unmarshal(req.Body, obj)
}
