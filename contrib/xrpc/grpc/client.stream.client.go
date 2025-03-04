package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/xrpc"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
)

var _ xrpc.ClientStreamClient = (*grpcClientStreamRequest)(nil)

type grpcClientStreamRequest struct {
	servicePath  string
	header       xtypes.SMap
	method       string
	streamClient grpcproto.GRPC_ClientStreamProcessClient
	onceLock     sync.Once
}

func (c *grpcClientStreamRequest) Send(obj any) error {
	var bodyBytes []byte
	switch t := obj.(type) {
	case []byte:
		bodyBytes = t
	case string:
		bodyBytes = bytesconv.StringToBytes(t)
	case *string:
		bodyBytes = bytesconv.StringToBytes(*t)
	default:
		bodyBytes, _ = json.Marshal(t)
	}

	return c.streamClient.Send(&grpcproto.Request{
		Body:    bodyBytes,
		Header:  c.header,
		Method:  c.method,
		Service: c.servicePath,
	})
}

func (c *Client) ClientStreamProcessor(ctx context.Context, processor xrpc.ClientStreamProcessor, opts *xrpc.Options) (body xrpc.Body, err error) {
	servicePath := c.reqPath.Path
	if len(opts.Query) > 0 {
		servicePath = fmt.Sprintf("%s?%s", servicePath, opts.Query)
	}
	grpcOpts := c.buildGrpcOpts(opts)

	clientStream, err := c.client.ClientStreamProcess(ctx, grpcOpts...)
	if err != nil {
		return xrpc.NewEmptyBody(), err
	}

	//发送服务分发数据信息
	err = clientStream.Send(&grpcproto.Request{
		Method:  opts.Method,
		Service: servicePath,
		Header:  opts.Header,
	})
	if err != nil {
		return xrpc.NewEmptyBody(), err
	}

	err = processor(&grpcClientStreamRequest{
		servicePath:  servicePath,
		header:       opts.Header,
		method:       opts.Method,
		streamClient: clientStream,
	})

	resp, err := clientStream.CloseAndRecv()
	if err != nil {
		return nil, err
	}
	return resp, err
}
