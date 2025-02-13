package grpc

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/xrpc"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
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
