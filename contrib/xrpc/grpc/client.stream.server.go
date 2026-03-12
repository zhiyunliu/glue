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

var _ xrpc.ServerStreamClient = (*grpcServerStreamRequest)(nil)

type grpcServerStreamRequest struct {
	servicePath  string
	header       engine.Header
	method       string
	streamClient grpcproto.GRPC_ServerStreamProcessClient
	onceLock     sync.Once
	RecvCount    int
}

func (c *grpcServerStreamRequest) Recv(obj any, opts ...xrpc.StreamRevcOption) (closed bool, err error) {
	opt := xrpc.StreamRecvOptions{
		Unmarshal: defaultUnmarshaler,
	}
	for _, o := range opts {
		o(&opt)
	}
	c.RecvCount++
	resp, err := c.streamClient.Recv()
	if err != nil {
		if err.Error() != "EOF" {
			err = fmt.Errorf("client.grpc.stream.Recv:%v", err)
			return
		}
		return true, nil
	}
	if obj == nil {
		return false, nil
	}
	return false, opt.Unmarshal(resp.Result, obj)
}

func (c *Client) ServerStreamProcessor(ctx context.Context, processor xrpc.ServerStreamProcessor, input any, opts *xrpc.Options) (err error) {
	servicePath := c.reqPath.Path
	if len(opts.Query) > 0 {
		servicePath = fmt.Sprintf("%s?%s", servicePath, opts.Query)
	}
	grpcOpts := c.buildGrpcOpts(opts)

	var bodyBytes []byte
	switch t := input.(type) {
	case []byte:
		bodyBytes = t
	case string:
		bodyBytes = bytesconv.StringToBytes(t)
	case *string:
		bodyBytes = bytesconv.StringToBytes(*t)
	default:
		bodyBytes, _ = json.Marshal(t)
	}

	req := &grpcproto.Request{
		Method:  opts.Method,
		Service: servicePath,
		Header:  opts.Header,
		Body:    bodyBytes,
	}

	serverStream, err := c.client.ServerStreamProcess(ctx, req, grpcOpts...)
	if err != nil {
		inerr := fmt.Errorf("ServerStream grpc://%s%s,BidirectionalStreamProcess,ServerStreamProcess:%w", c.reqPath.Host, servicePath, err)
		err = errors.Clone(constants.ErrRemoteRequest, errors.WithInnerErr(inerr))
		return err
	}

	err = processor(ctx, &grpcServerStreamRequest{
		servicePath:  servicePath,
		header:       opts.Header,
		method:       opts.Method,
		streamClient: serverStream,
	})
	if err != nil {
		inerr := fmt.Errorf("ServerStream grpc://%s%s,BidirectionalStreamProcess,processor:%w", c.reqPath.Host, servicePath, err)
		err = errors.Clone(constants.ErrRemoteRequest, errors.WithInnerErr(inerr))
	}
	return err
}
