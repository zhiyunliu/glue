package main

import (
	"time"

	"github.com/zhiyunliu/velocity"
	"github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/server/api"
)

func main() {
	apiSrv := api.New("xx")
	//mqcSrv := mqc.New("bb")

	apiSrv.Handle("/demo", func(ctx context.Context) interface{} {
		ctx.Log().Debug("xxx")
		time.Sleep(time.Second)
		return ""
	})

	app := velocity.NewApp(velocity.Server(apiSrv))
	app.Start()
}
