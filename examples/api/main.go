package main

import (
	"fmt"

	"github.com/zhiyunliu/golibs/xtypes"
	"github.com/zhiyunliu/velocity"
	"github.com/zhiyunliu/velocity/context"
	_ "github.com/zhiyunliu/velocity/contrib/registry/nacos"
	"github.com/zhiyunliu/velocity/errors"
	"github.com/zhiyunliu/velocity/server/api"
)

func main() {
	apiSrv := api.New("xx")
	//mqcSrv := mqc.New("bb")

	apiSrv.Handle("/demo", func(ctx context.Context) interface{} {
		ctx.Log().Debug("demo")
		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})

	apiSrv.Handle("/error", func(ctx context.Context) interface{} {
		ctx.Log().Debug("error")
		return errors.New(300, "xxx")
	})

	apiSrv.Handle("/panic", func(ctx context.Context) interface{} {
		ctx.Log().Debug("panic")
		panic(fmt.Errorf("xx i am panic"))
		return fmt.Errorf("xxx")
	})

	app := velocity.NewApp(velocity.Server(apiSrv))
	app.Start()
}
