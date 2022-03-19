package main

import (
	"github.com/zhiyunliu/velocity"
	"github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/server/api"
	"github.com/zhiyunliu/velocity/server/mqc"
)

func main() {
	apiSrv := api.New("xx")
	mqcSrv := mqc.New("bb")

	apiSrv.Handle("/demo", func(ctx context.Context) interface{} {
		ctx.Log().Debug("xxx")
		return ""
	})

	app := velocity.NewApp(velocity.Server(apiSrv, mqcSrv))
	app.Start()
}
