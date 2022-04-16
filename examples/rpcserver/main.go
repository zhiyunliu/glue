package main

import (
	"github.com/zhiyunliu/gel"
	_ "github.com/zhiyunliu/gel/contrib/config/nacos"
	_ "github.com/zhiyunliu/gel/contrib/queue/redis"
	_ "github.com/zhiyunliu/gel/contrib/registry/consul"
	_ "github.com/zhiyunliu/gel/contrib/registry/nacos"

	"github.com/zhiyunliu/gel/examples/cronserver/demos"
	"github.com/zhiyunliu/gel/server/rpc"
)

func main() {
	rcpSrv := rpc.New("")

	rcpSrv.Handle("/demo", &demos.Fulldemo{})

	app := gel.NewApp(gel.Server(rcpSrv))

	app.Start()
}
