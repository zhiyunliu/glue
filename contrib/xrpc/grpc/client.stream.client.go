package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/xrpc"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
	"golang.org/x/sync/errgroup"
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

func buildDefaultStreamProcess(input any) (processor xrpc.BidirectionalStreamProcessor, err error) {
	refval := reflect.ValueOf(input)
	if refval.IsNil() {
		return nil, fmt.Errorf("input is nil")
	}

	refType := refval.Type()
	if refType.Kind() != reflect.Array &&
		refType.Kind() != reflect.Slice {
		return nil, fmt.Errorf("input is not array or slice")
	}

	//[]byte数组
	if refType.Kind() == reflect.Slice &&
		refType.Elem().Kind() == reflect.Uint8 {
		return nil, fmt.Errorf("input is []byte array")
	}

	return func(client xrpc.BidirectionalStreamClient) error {
		errGroup := errgroup.Group{}
		//调用grpc服务
		errGroup.Go(func() error {
			for i, valLen := 0, refval.Len(); i < valLen; i++ {
				item := refval.Index(i).Interface()
				if err := client.Send(item); err != nil {
					return err
				}
			}
			return client.CloseSend()
		})

		errGroup.Go(func() error {
			for {
				closed, err := client.Recv(nil)
				if err != nil || closed {
					return err
				}
			}
		})
		err := errGroup.Wait()
		return err
	}, nil
}
