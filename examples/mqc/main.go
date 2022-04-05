package main

import (
	"github.com/zhiyunliu/gel"
	_ "github.com/zhiyunliu/gel/contrib/queue/redis"
	"github.com/zhiyunliu/gel/server/mqc"
)

func main() {
	mqcSrv := mqc.New("")

	mqcSrv.Handle("yy", &demo{})

	app := gel.NewApp(gel.Server(mqcSrv))

	app.Start()
}
