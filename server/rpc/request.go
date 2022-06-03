package rpc

import (
	"bytes"
	"context"
	sctx "context"
	"encoding/json"
	"io"

	"github.com/zhiyunliu/gel/constants"
	"github.com/zhiyunliu/gel/contrib/xrpc/grpc/grpcproto"
	"google.golang.org/grpc/peer"

	"github.com/zhiyunliu/gel/contrib/alloter"
)

var _ alloter.IRequest = (*Request)(nil)

//Request 处理任务请求
type Request struct {
	ctx    sctx.Context
	rpcReq *grpcproto.Request
	method string
	params map[string]string
	header map[string]string
	body   cbody
}

//NewRequest 构建任务请求
func NewRequest(ctx context.Context, rpcReq *grpcproto.Request) (r *Request, err error) {

	r = &Request{
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

//GetName 获取任务名称
func (m *Request) GetName() string {
	return m.rpcReq.Service
}

//GetService 服务名
func (m *Request) GetService() string {
	return m.rpcReq.Service
}

//GetMethod 方法名
func (m *Request) GetMethod() string {
	return m.method
}

func (m *Request) Params() map[string]string {
	return m.params
}

func (m *Request) GetHeader() map[string]string {
	return m.header
}

func (m *Request) Body() []byte {
	return m.body
}

func (m *Request) GetRemoteAddr() string {
	if p, ok := peer.FromContext(m.ctx); ok {
		return p.Addr.String()
	}
	return m.header[constants.HeaderRemoteHeader]
}

func (m *Request) Context() sctx.Context {
	return m.ctx
}
func (m *Request) WithContext(ctx sctx.Context) alloter.IRequest {
	m.ctx = ctx
	return m
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
