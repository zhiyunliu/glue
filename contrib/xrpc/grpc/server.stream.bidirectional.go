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

var _ alloter.IRequest = (*bidirectionalStreamRequest)(nil)
var _ xrpc.BidirectionalStreamRequest = (*bidirectionalStreamRequest)(nil)

// Request 处理任务请求
type bidirectionalStreamRequest struct {
	ctx      sctx.Context
	firstReq *grpcproto.Request
	url      *url.URL
	method   string
	params   map[string]string
	header   map[string]string
	stream   grpcproto.GRPC_BidirectionalStreamProcessServer
}

// newStreamRequest 构建任务请求
func newBidirectionalStreamRequest(stream grpcproto.GRPC_BidirectionalStreamProcessServer) (r *bidirectionalStreamRequest, err error) {

	// 接收第一个请求(服务分发数据信息)
	firstReq, err := stream.Recv()
	if err != nil {
		err = fmt.Errorf("server.grpc.stream.Recv(first):%v", err)
		return
	}

	r = &bidirectionalStreamRequest{
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
func (m *bidirectionalStreamRequest) GetName() string {
	return m.firstReq.Service
}

// GetService 服务名()
func (m *bidirectionalStreamRequest) GetService() string {
	return m.firstReq.Service
}

// GetService 服务名()
func (m *bidirectionalStreamRequest) GetURL() *url.URL {
	if m.url == nil {
		m.url, _ = url.Parse(m.firstReq.Service)
	}
	return m.url
}

// GetMethod 方法名
func (m *bidirectionalStreamRequest) GetMethod() string {
	return m.method
}

func (m *bidirectionalStreamRequest) Params() map[string]string {
	return m.params
}

func (m *bidirectionalStreamRequest) GetHeader() engine.Header {
	return m.header
}

func (m *bidirectionalStreamRequest) Body() []byte {
	return []byte{}
}

func (m *bidirectionalStreamRequest) GetRemoteAddr() string {
	if p, ok := peer.FromContext(m.ctx); ok {
		return p.Addr.String()
	}
	return m.header[constants.HeaderRemoteHeader]
}

func (m *bidirectionalStreamRequest) Context() sctx.Context {
	return m.ctx
}
func (m *bidirectionalStreamRequest) WithContext(ctx sctx.Context) {
	m.ctx = ctx
}

//// StreamRequest ////

func (m *bidirectionalStreamRequest) Send(obj any) (err error) {
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

// Recv 反序列化请求体
func (m *bidirectionalStreamRequest) Recv(obj any, opts ...xrpc.StreamRevcOption) (closed bool, err error) {
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
