package grpc

import (
	"fmt"

	"context"

	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/errors/constants"
	"github.com/zhiyunliu/glue/xrpc"
)

func (c *Client) clientRequest(ctx context.Context, o *xrpc.Options, bodyBytes []byte) (resp *grpcproto.Response, err error) {
	servicePath := c.reqPath.Path
	if len(o.Query) > 0 {
		servicePath = fmt.Sprintf("%s?%s", servicePath, o.Query)
	}

	req := &grpcproto.Request{
		Method:  o.Method, //借用http的method
		Service: servicePath,
		Header:  o.Header,
		Body:    bodyBytes,
	}

	if req.Header == nil {
		req.Header = make(map[string]string)
	}

	// 调用grpc服务
	resp, err = c.client.Process(ctx, req, c.buildGrpcOpts(o)...)
	// 处理响应
	if err != nil {

		inerr := fmt.Errorf("Normal grpc://%s%s,Process,error:%w", c.reqPath.Host, servicePath, err)
		err = errors.Clone(constants.ErrRemoteRequest, errors.WithInnerErr(inerr))
		return
	}
	return
}
