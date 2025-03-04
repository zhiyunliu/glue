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

var _ xrpc.BidirectionalStreamClient = (*grpcBidirectionalClientStreamRequest)(nil)

type grpcBidirectionalClientStreamRequest struct {
	servicePath  string
	header       xtypes.SMap
	method       string
	streamClient grpcproto.GRPC_BidirectionalStreamProcessClient
	onceLock     sync.Once
}

func (c *grpcBidirectionalClientStreamRequest) Recv(obj any, opts ...xrpc.StreamRevcOption) (closed bool, err error) {
	opt := xrpc.StreamRecvOptions{
		Unmarshal: defaultUnmarshaler,
	}
	for _, o := range opts {
		o(&opt)
	}
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
func (c *grpcBidirectionalClientStreamRequest) Send(obj any) error {
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

func (c *grpcBidirectionalClientStreamRequest) CloseSend() (err error) {
	c.onceLock.Do(func() {
		if c.streamClient != nil {
			err = c.streamClient.CloseSend()
		}
	})
	return err
}

func (c *Client) BidirectionalStreamProcessor(ctx context.Context, processor xrpc.BidirectionalStreamProcessor, opts *xrpc.Options) error {
	servicePath := c.reqPath.Path
	if len(opts.Query) > 0 {
		servicePath = fmt.Sprintf("%s?%s", servicePath, opts.Query)
	}
	grpcOpts := c.buildGrpcOpts(opts)

	steamClient, err := c.client.BidirectionalStreamProcess(ctx, grpcOpts...)
	if err != nil {
		return err
	}
	//发送服务分发数据信息
	err = steamClient.Send(&grpcproto.Request{
		Method:  opts.Method,
		Service: servicePath,
		Header:  opts.Header,
	})
	if err != nil {
		return err
	}
	err = processor(&grpcBidirectionalClientStreamRequest{
		servicePath:  servicePath,
		header:       opts.Header,
		method:       opts.Method,
		streamClient: steamClient,
	})
	return err
}
