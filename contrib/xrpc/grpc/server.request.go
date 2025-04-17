package grpc

import (
	"bytes"
	sctx "context"
	"encoding/json"
	"io"
	"net/url"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/engine"
	"google.golang.org/grpc/peer"

	"github.com/zhiyunliu/alloter"
)

var _ alloter.IRequest = (*serverRequest)(nil)

// Request 处理任务请求
type serverRequest struct {
	ctx    sctx.Context
	rpcReq *grpcproto.Request
	url    *url.URL
	method string
	params map[string]string
	header map[string]string
	body   cbody
}

// NewRequest 构建任务请求
func newServerRequest(ctx sctx.Context, rpcReq *grpcproto.Request) (r *serverRequest, err error) {

	r = &serverRequest{
		rpcReq: rpcReq,
		method: rpcReq.Method,
		header: rpcReq.Header,
		params: make(map[string]string),
	}
	if r.header == nil {
		r.header = map[string]string{}
	}
	r.body = cbody(rpcReq.Body)

	r.ctx = ctx

	return r, nil
}

// GetName 获取任务名称
func (m *serverRequest) GetName() string {
	return m.rpcReq.Service
}

// GetService 服务名()
func (m *serverRequest) GetService() string {
	return m.rpcReq.Service
}

// GetService 服务名()
func (m *serverRequest) GetURL() *url.URL {
	if m.url == nil {
		m.url, _ = url.Parse(m.rpcReq.Service)
	}
	return m.url
}

// GetMethod 方法名
func (m *serverRequest) GetMethod() string {
	return m.method
}

func (m *serverRequest) Params() map[string]string {
	return m.params
}

func (m *serverRequest) GetHeader() engine.Header {
	return m.header
}

func (m *serverRequest) Body() []byte {
	return m.body
}

func (m *serverRequest) GetRemoteAddr() string {
	if p, ok := peer.FromContext(m.ctx); ok {
		return p.Addr.String()
	}
	return m.header[constants.HeaderRemoteHeader]
}

func (m *serverRequest) Context() sctx.Context {
	return m.ctx
}
func (m *serverRequest) WithContext(ctx sctx.Context) {
	m.ctx = ctx
}

type Body interface {
	io.Reader
	Scan(obj interface{}) error
}

type cbody []byte

func (b cbody) Read(p []byte) (n int, err error) {
	return bytes.NewReader(b).Read(p)
}

func (b cbody) Scan(obj interface{}) error {
	return json.Unmarshal(b, obj)
}
