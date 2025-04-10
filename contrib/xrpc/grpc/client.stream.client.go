package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/xrpc"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var _ xrpc.ClientStreamClient = (*grpcClientStreamRequest)(nil)

type grpcClientStreamRequest struct {
	servicePath  string
	header       xtypes.SMap
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
		return xrpc.NewEmptyBody(), err
	}

	req := &grpcproto.Request{
		Method:  opts.Method,
		Service: servicePath,
		Header:  opts.Header,
	}
	ctx, span := GetStreamSpanFromContext(ctx, req)
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.type", "clientstream"),
	)
	//发送服务分发数据信息
	err = clientStream.Send(req)
	if err != nil {
		return xrpc.NewEmptyBody(), err
	}

	clientStreamRequest := &grpcClientStreamRequest{
		servicePath:  servicePath,
		header:       opts.Header,
		method:       opts.Method,
		streamClient: clientStream,
	}

	err = processor(ctx, clientStreamRequest)

	resp, err := clientStream.CloseAndRecv()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(
		attribute.Int("rpc.stream.send", clientStreamRequest.SendCount),
	)
	span.SetAttributes(
		attribute.Int("rpc.response.status_code", resp.StatusCode()),
		attribute.Int("rpc.response.body.size", len(resp.Result)),
	)
	if resp.Status >= http.StatusBadRequest {
		span.SetStatus(codes.Error, http.StatusText(int(resp.Status)))
	}
	return resp, err
}
