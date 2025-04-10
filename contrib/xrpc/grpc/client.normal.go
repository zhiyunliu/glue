package grpc

import (
	"fmt"
	"net/http"

	"context"

	"github.com/zhiyunliu/glue/contrib/xrpc/grpc/grpcproto"
	"github.com/zhiyunliu/glue/xrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

	ctx, span := GetNormalSpanFromContext(ctx, req)
	defer span.End()

	span.SetAttributes(
		attribute.String("rpc.type", "normal"),
		attribute.Int("rpc.request.body.size", len(bodyBytes)),
	)
	// 调用grpc服务
	resp, err = c.client.Process(ctx, req, c.buildGrpcOpts(o)...)
	// 处理响应
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	span.SetAttributes(
		attribute.Int("rpc.response.status_code", resp.StatusCode()),
		attribute.Int("rpc.response.body.size", len(resp.Result)),
	)
	if resp.Status >= http.StatusBadRequest {
		span.SetStatus(codes.Error, http.StatusText(int(resp.Status)))
	}
	return
}
