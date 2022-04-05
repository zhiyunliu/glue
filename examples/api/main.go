package main

import (
	"fmt"

	"github.com/zhiyunliu/gel"
	"github.com/zhiyunliu/gel/context"
	_ "github.com/zhiyunliu/gel/contrib/config/nacos"
	_ "github.com/zhiyunliu/gel/contrib/registry/nacos"
	"github.com/zhiyunliu/gel/errors"
	"github.com/zhiyunliu/gel/server/api"
	"github.com/zhiyunliu/golibs/xtypes"
)

func main() {
	apiSrv := api.New("")
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

	app := gel.NewApp(gel.Server(apiSrv))
	app.Start()
}
