package grpc

import (
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

var _ xrpc.StreamClient = (*grpcClientStreamRequest)(nil)

type grpcClientStreamRequest struct {
	servicePath  string
	header       xtypes.SMap
	method       string
	streamClient grpcproto.GRPC_StreamProcessClient
	onceLock     sync.Once
}

func (c *grpcClientStreamRequest) Recv(obj any, opts ...xrpc.StreamRevcOption) (closed bool, err error) {
	opt := xrpc.StreamRecvOptions{
		Unmarshal: unmarshaler,
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

func (c *grpcClientStreamRequest) CloseSend() (err error) {
	c.onceLock.Do(func() {
		if c.streamClient != nil {
			err = c.streamClient.CloseSend()
		}
	})
	return err
}

func buildDefaultStreamProcess(input any) (processor xrpc.StreamProcessor, err error) {
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

	return func(client xrpc.StreamClient) error {
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
