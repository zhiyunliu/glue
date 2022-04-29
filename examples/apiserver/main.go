package main

import (
	"fmt"

	"github.com/zhiyunliu/gel"
	"github.com/zhiyunliu/gel/context"
	_ "github.com/zhiyunliu/gel/contrib/cache/redis"
	_ "github.com/zhiyunliu/gel/contrib/config/consul"
	_ "github.com/zhiyunliu/gel/contrib/config/nacos"
	_ "github.com/zhiyunliu/gel/contrib/queue/redis"
	_ "github.com/zhiyunliu/gel/contrib/registry/nacos"
	_ "github.com/zhiyunliu/gel/contrib/xdb/mysql"
	_ "github.com/zhiyunliu/gel/contrib/xdb/sqlserver"
	"github.com/zhiyunliu/gel/errors"
	"github.com/zhiyunliu/gel/examples/apiserver/demos"
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
	})

	apiSrv.Handle("/db", demos.NewDb())
	apiSrv.Handle("/cache", demos.NewCache())
	apiSrv.Handle("/queue", demos.NewQueue())
	apiSrv.Handle("/log", demos.NewLogDemo())
	apiSrv.Handle("/rpc", demos.NewGrpcDemo())

	app := gel.NewApp(gel.Server(apiSrv))
	app.Start()
}
