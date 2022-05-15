package main

import (
	"net/http"
	"time"

	"github.com/zhiyunliu/gel"
	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/gel/examples/compositeserver/handles"
	"github.com/zhiyunliu/gel/transport"
	"github.com/zhiyunliu/gel/xhttp"
	"github.com/zhiyunliu/gel/xrpc"

	"github.com/zhiyunliu/gel/server/api"
	"github.com/zhiyunliu/gel/server/cron"
	"github.com/zhiyunliu/gel/server/mqc"
	"github.com/zhiyunliu/gel/server/rpc"
	"github.com/zhiyunliu/golibs/xtypes"
)

func init() {
	srvOpt := gel.Server(
		apiserver(),
		mqcserver(),
		cronserver(),
		rpcserver(),
	)
	opts = append(opts, srvOpt, gel.LogConcurrency(1))
}

func apiserver() transport.Server {
	apiSrv := api.New("apiserver")
	apiSrv.Handle("/log", handles.NewLogDemo())
	apiSrv.Handle("/demoapi", func(ctx context.Context) interface{} {
		ctx.Log().Debug("demo")

		body, err := gel.RPC().GetRPC().Swap(ctx, "grpc://compositeserver/demorpc", xrpc.WithWaitForReady(false))
		if err != nil {
			ctx.Log().Error("gel.RPC().GetRPC().Swap:", err)
		}
		ctx.Log().Debug(string(body.GetResult()))
		ctx.Log().Debug(body.GetHeader())
		ctx.Log().Debug(body.GetStatus())
		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})
	return apiSrv
}

func mqcserver() transport.Server {
	mqcSrv := mqc.New("mqcserver")

	mqcSrv.Handle("/demomqc", func(ctx context.Context) interface{} {
		ctx.Log().Debug("demomqc")
		body, err := gel.Http().GetHttp().Swap(ctx, "http://compositeserver/demoapi", xhttp.WithMethod(http.MethodPost))
		if err != nil {
			ctx.Log().Error("gel.Http().GetHttp().Swap:", err)
		}
		ctx.Log().Debug(string(body.GetResult()))
		ctx.Log().Debug(body.GetHeader())
		ctx.Log().Debug(body.GetStatus())

		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})

	return mqcSrv
}

func rpcserver() transport.Server {
	rpcSrv := rpc.New("rpcserver")

	rpcSrv.Handle("/demorpc", func(ctx context.Context) interface{} {
		ctx.Log().Debug("demorpc")
		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})

	return rpcSrv
}

func cronserver() transport.Server {
	cronSrv := cron.New("cronserver")

	cronSrv.Handle("/democron", func(ctx context.Context) interface{} {
		ctx.Log().Debug("democron")

		gel.Queue().GetQueue("default").Send(ctx, "xx.xx.xx", map[string]interface{}{
			"a": time.Now().Unix(),
		})

		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})
	return cronSrv
}
