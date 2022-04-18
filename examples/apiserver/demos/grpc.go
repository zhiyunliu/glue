package demos

import (
	"strconv"

	"github.com/zhiyunliu/gel"
	"github.com/zhiyunliu/gel/context"
	_ "github.com/zhiyunliu/gel/contrib/xrpc/grpc"
	"github.com/zhiyunliu/gel/xrpc"
)

type GrpcDemo struct{}

func NewGrpcDemo() *GrpcDemo {
	return &GrpcDemo{}
}

func (d *GrpcDemo) RequestHandle(ctx context.Context) interface{} {

	wfr := ctx.Request().Query().Get("wfr")
	wfrv, _ := strconv.ParseBool(wfr)

	client := gel.RPC().GetRPC("default")
	body, err := client.Request(ctx.Context(), "grpc://rpcserver/demo", map[string]interface{}{
		"body-a": "1",
		"body-b": 2,
		"body-c": struct {
			A string
		}{
			A: "s-1",
		},
	}, xrpc.WithXRequestID("aaa"), xrpc.WithWaitForReady(wfrv))
	if err != nil {
		ctx.Log().Error(err)
		return err
	}

	ctx.Log().Info("body.GetHeader", body.GetHeader())
	ctx.Log().Info("body.GetStatus", body.GetStatus())
	ctx.Log().Info("body.GetResult", string(body.GetResult()))
	return string(body.GetResult())
}

func (d *GrpcDemo) SwapHandle(ctx context.Context) interface{} {
	client := gel.RPC().GetRPC("")
	body, err := client.Swap(ctx, "grpc://rpcserver/demo", xrpc.WithXRequestID("aaa"))
	if err != nil {
		ctx.Log().Error(err)
		return err
	}

	ctx.Log().Info("body.GetHeader", body.GetHeader())
	ctx.Log().Info("body.GetStatus", body.GetStatus())
	ctx.Log().Info("body.GetResult", string(body.GetResult()))
	return "success"
}
