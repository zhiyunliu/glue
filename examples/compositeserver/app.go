package main

import (
	"time"

	"github.com/zhiyunliu/gel"
	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/gel/examples/compositeserver/handles"

	"github.com/zhiyunliu/gel/server/api"
	"github.com/zhiyunliu/gel/server/mqc"
	"github.com/zhiyunliu/gel/server/rpc"
	"github.com/zhiyunliu/golibs/xtypes"
)

func init() {
	apiserver()
	mqcserver()
	rpcserver()
	cronserver()
}

func apiserver() {
	apiSrv := api.New("apiserver")
	apiSrv.Handle("/log", handles.NewLogDemo())
	apiSrv.Handle("/demo", func(ctx context.Context) interface{} {
		ctx.Log().Debug("demo")
		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})

	opts = append(opts, gel.Server(apiSrv))
}

func mqcserver() {
	mqcSrv := mqc.New("mqcserver")

	mqcSrv.Handle("/demomqc", func(ctx context.Context) interface{} {
		ctx.Log().Debug("demomqc")
		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})

	opts = append(opts, gel.Server(mqcSrv))
}

func rpcserver() {
	rpcSrv := rpc.New("rpcserver")

	rpcSrv.Handle("/demorpc", func(ctx context.Context) interface{} {
		ctx.Log().Debug("demorpc")
		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})

	opts = append(opts, gel.Server(rpcSrv))
}

func cronserver() {
	cronSrv := rpc.New("cronserver")

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

	opts = append(opts, gel.Server(cronSrv))
}
