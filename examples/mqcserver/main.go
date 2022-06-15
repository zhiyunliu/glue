package main

import (
	gel "github.com/zhiyunliu/glue"
	_ "github.com/zhiyunliu/glue/contrib/config/nacos"
	_ "github.com/zhiyunliu/glue/contrib/queue/redis"
	_ "github.com/zhiyunliu/glue/contrib/registry/nacos"
	"github.com/zhiyunliu/glue/examples/mqcserver/demos"
	"github.com/zhiyunliu/glue/server/mqc"
)

func main() {
	mqcSrv := mqc.New("")

	mqcSrv.Handle("yy", &demos.Orgdemo{})
	mqcSrv.Handle("key", &demos.Fulldemo{})

	app := gel.NewApp(gel.Server(mqcSrv))

	app.Start()
}
