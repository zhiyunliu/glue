package demos

import (
	"github.com/zhiyunliu/gel"
	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/gel/xgrpc"
)

type GrpcDemo struct{}

func NewGrpcDemo() *GrpcDemo {
	return &GrpcDemo{}
}

func (d *GrpcDemo) RequestHandle(ctx context.Context) interface{} {

	client := gel.RPC().GetRPC("default")
	body, err := client.Request(ctx.Context(), "grpc://rpcserver/demo", map[string]interface{}{}, xgrpc.WithTraceID("aaa"))
	if err != nil {
		ctx.Log().Error(err)
		return err
	}

	ctx.Log().Info("body.GetHeader", body.GetHeader())
	ctx.Log().Info("body.GetStatus", body.GetStatus())
	ctx.Log().Info("body.GetResult", string(body.GetResult()))
	return "success"
}

func (d *GrpcDemo) SwapHandle(ctx context.Context) interface{} {
	client := gel.RPC().GetRPC("")
	body, err := client.Swap(ctx, "grpc://rpcserver/demo", xgrpc.WithTraceID("aaa"))
	if err != nil {
		ctx.Log().Error(err)
		return err
	}

	ctx.Log().Info("body.GetHeader", body.GetHeader())
	ctx.Log().Info("body.GetStatus", body.GetStatus())
	ctx.Log().Info("body.GetResult", string(body.GetResult()))
	return "success"
}
