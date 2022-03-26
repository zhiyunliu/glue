package main

import (
	"github.com/zhiyunliu/velocity"
	_ "github.com/zhiyunliu/velocity/contrib/queue/redis"
	"github.com/zhiyunliu/velocity/server/mqc"
)

func main() {
	mqcSrv := mqc.New("")

	mqcSrv.Handle("xx.xx.xx", &demo{})

	app := velocity.NewApp(velocity.Server(mqcSrv))

	app.Start()
}
