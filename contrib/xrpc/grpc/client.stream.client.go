package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/errors/constants"
	"github.com/zhiyunliu/glue/xrpc"
	"github.com/zhiyunliu/golibs/bytesconv"
)

var _ xrpc.ClientStreamClient = (*grpcClientStreamRequest)(nil)

type grpcClientStreamRequest struct {
	servicePath  string
	header       engine.Header
	method       string
	streamClient grpcproto.GRPC_ClientStreamProcessClient
	onceLock     sync.Once
	SendCount    int
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
	c.SendCount++
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
		inerr := fmt.Errorf("ClientStream grpc://%s%s,BidirectionalStreamProcess,ClientStreamProcess:%w", c.reqPath.Host, servicePath, err)
		err = errors.Clone(constants.ErrRemoteRequest, errors.WithInnerErr(inerr))
		return xrpc.NewEmptyBody(), err
	}

	//发送服务分发数据信息
	req := &grpcproto.Request{
		Method:  opts.Method,
		Service: servicePath,
		Header:  opts.Header,
	}

	//发送服务分发数据信息
	err = clientStream.Send(req)
	if err != nil {
		inerr := fmt.Errorf("ClientStream grpc://%s%s,BidirectionalStreamProcess,Send:%w", c.reqPath.Host, servicePath, err)
		err = errors.Clone(constants.ErrRemoteRequest, errors.WithInnerErr(inerr))
		return xrpc.NewEmptyBody(), err
	}

	clientStreamRequest := &grpcClientStreamRequest{
		servicePath:  servicePath,
		header:       opts.Header,
		method:       opts.Method,
		streamClient: clientStream,
	}

	err = processor(ctx, clientStreamRequest)
	if err != nil {

		inerr := fmt.Errorf("ClientStream grpc://%s%s,BidirectionalStreamProcess,processor:%w", c.reqPath.Host, servicePath, err)
		err = errors.Clone(constants.ErrRemoteRequest, errors.WithInnerErr(inerr))
		return nil, err
	}
	resp, err := clientStream.CloseAndRecv()
	if err != nil {
		inerr := fmt.Errorf("ClientStream grpc://%s%s,BidirectionalStreamProcess,CloseAndRecv:%w", c.reqPath.Host, servicePath, err)
		err = errors.Clone(constants.ErrRemoteRequest, errors.WithInnerErr(inerr))
		return nil, err
	}
	return resp, err
}
