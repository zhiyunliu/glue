package main

import (
	"math/rand"
	"time"

	sctx "context"

	"github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/glue/context"
	_ "github.com/zhiyunliu/glue/contrib/cache/redis"
	_ "github.com/zhiyunliu/glue/contrib/config/consul"
	_ "github.com/zhiyunliu/glue/contrib/config/nacos"
	_ "github.com/zhiyunliu/glue/contrib/queue/redis"
	_ "github.com/zhiyunliu/glue/contrib/registry/nacos"
	_ "github.com/zhiyunliu/glue/contrib/xdb/mysql"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/server/api"
	"github.com/zhiyunliu/golibs/xlog"

	//_ "github.com/zhiyunliu/glue/contrib/xdb/oracle"
	_ "github.com/zhiyunliu/glue/contrib/dlocker/redis"
	_ "github.com/zhiyunliu/glue/contrib/xdb/postgres"
	_ "github.com/zhiyunliu/glue/contrib/xdb/sqlite"
	_ "github.com/zhiyunliu/glue/contrib/xdb/sqlserver"
)

type demo struct{}

func (d *demo) InfoHandle(ctx context.Context) interface{} {
	return "success"
}
func (d *demo) DetailHandle(ctx context.Context) interface{} {
	return map[string]interface{}{
		"detail": "i am demo",
	}
}

type body struct {
	Seq string `json:"seq"`
}

func main() {
	rand.Seed(time.Now().UnixMilli())
	apiSrv := api.New("apiserver", api.WithServiceName("demo"), api.Log(log.WithRequest(), log.WithResponse()))
	apiSrv.Handle("/demo/origin", func(ctx context.Context) interface{} {
		slp := rand.Intn(6)
		time.Sleep(time.Second * time.Duration(slp))
		ver := ctx.Request().Query().Get("ver")
		b := &body{}
		ctx.Bind(b)
		return map[string]interface{}{
			"v":   ver,
			"b":   b.Seq,
			"slp": slp,
		}
	})
	apiSrv.Handle("/demo/struct", &demo{})
	apiSrv.Handle("/demo", func(ctx context.Context) interface{} {
		return map[string]interface{}{
			"a": "1",
		}
	})
	apiSrv.Handle("/log", func(ctx context.Context) interface{} {
		return xlog.Stats()
	})

	app := glue.NewApp(glue.Server(apiSrv), glue.StartedHook(func(ctx sctx.Context) error {
		log.Debug("global.Config:", global.Config)
		xx := &XX{}
		global.Config.Scan(xx)
		log.Debugf("XX:%+v", xx)
		return nil
	}), glue.StartingHook(func(ctx sctx.Context) error {
		log.Debug("global.Config.start:", global.Config)
		return nil
	}))
	app.Start()
}

type XX struct {
	A string `json:"a"`
	B int    `json:"b"`
}
