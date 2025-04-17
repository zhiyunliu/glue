package grpc

import (
	sctx "context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/zhiyunliu/alloter"
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/xrpc"
	"google.golang.org/grpc/peer"
)

var _ alloter.IRequest = (*serverStreamRequest)(nil)
var _ xrpc.ServerStreamRequest = (*serverStreamRequest)(nil)

// Request 处理任务请求
type serverStreamRequest struct {
	ctx       sctx.Context
	url       *url.URL
	service   string
	method    string
	dataBytes []byte
	params    map[string]string
	header    map[string]string
	stream    grpcproto.GRPC_ServerStreamProcessServer
}

// newStreamRequest 构建任务请求
func newServerStreamRequest(request *grpcproto.Request, stream grpcproto.GRPC_ServerStreamProcessServer) (r *serverStreamRequest, err error) {

	r = &serverStreamRequest{
		stream:    stream,
		service:   request.Service,
		method:    request.Method,
		header:    request.Header,
		dataBytes: request.Body,
		params:    make(map[string]string),
	}
	if r.header == nil {
		r.header = map[string]string{}
	}
	r.ctx = stream.Context()
	return r, nil
}

// GetName 获取任务名称
func (m *serverStreamRequest) GetName() string {
	return m.service
}

// GetService 服务名()
func (m *serverStreamRequest) GetService() string {
	return m.service
}

// GetService 服务名()
func (m *serverStreamRequest) GetURL() *url.URL {
	if m.url == nil {
		m.url, _ = url.Parse(m.service)
	}
	return m.url
}

// GetMethod 方法名
func (m *serverStreamRequest) GetMethod() string {
	return m.method
}

func (m *serverStreamRequest) Params() map[string]string {
	return m.params
}

func (m *serverStreamRequest) GetHeader() engine.Header {
	return m.header
}

func (m *serverStreamRequest) Body() []byte {
	return m.dataBytes
}

func (m *serverStreamRequest) GetRemoteAddr() string {
	if p, ok := peer.FromContext(m.ctx); ok {
		return p.Addr.String()
	}
	return m.header[constants.HeaderRemoteHeader]
}

func (m *serverStreamRequest) Context() sctx.Context {
	return m.ctx
}
func (m *serverStreamRequest) WithContext(ctx sctx.Context) {
	m.ctx = ctx
}

func (m *serverStreamRequest) Send(obj any) (err error) {
	respObj := &grpcproto.Response{
		Status: 200,
	}
	// 处理选项
	if entity, ok := obj.(engine.ResponseEntity); ok {
		respObj.Headers = entity.Header()
		respObj.Status = int32(entity.StatusCode())
		respObj.Result, err = entity.Body()
		if err != nil {
			return err
		}
	} else {
		respObj.Result, err = json.Marshal(obj)
		if err != nil {
			err = fmt.Errorf("server.grpc.stream.Marshal:%v", err)
			return err
		}
		respObj.Headers = map[string]string{
			constants.ContentTypeName: constants.ContentTypeApplicationJSON,
		}
	}
	err = m.stream.Send(respObj)
	if err != nil {
		err = fmt.Errorf("server.grpc.stream.Send:%v", err)
		return err
	}
	return nil
}
