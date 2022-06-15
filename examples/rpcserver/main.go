package main

import (
	gel "github.com/zhiyunliu/glue"
	_ "github.com/zhiyunliu/glue/contrib/config/nacos"
	_ "github.com/zhiyunliu/glue/contrib/queue/redis"
	_ "github.com/zhiyunliu/glue/contrib/registry/consul"
	_ "github.com/zhiyunliu/glue/contrib/registry/nacos"

	"github.com/zhiyunliu/glue/examples/rpcserver/demos"
	"github.com/zhiyunliu/glue/server/rpc"
)

func main() {
	rcpSrv := rpc.New("")

	rcpSrv.Handle("/demo", &demos.Fulldemo{})

	app := gel.NewApp(gel.Server(rcpSrv))

	app.Start()
}
