package main

import (
	"time"

	"github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/glue/context"
	_ "github.com/zhiyunliu/glue/contrib/config/nacos"   //配置中心
	_ "github.com/zhiyunliu/glue/contrib/registry/nacos" //注册中心
	"github.com/zhiyunliu/glue/log"

	"github.com/zhiyunliu/glue/server/rpc"
)

func main() {
	rcpSrv := rpc.New("payserver")

	rcpSrv.Handle("/demo", func(ctx context.Context) interface{} {

		//使用当前请求的session 打印日志。在一个请求中的所有日志都有相同的sessionid
		ctx.Log().Infof("rcpSrv.demo:%s", time.Now().Format("2006-01-02 15:04:05"))

		//使用系统session打印日志。
		log.Debug("debug")
		log.Debugf("debug:%s", "debug")
		log.Info("Info")
		log.Infof("Info:%s", "Info")
		log.Warn("Warn")
		log.Warnf("Warn:%s", "Warn")
		log.Error("Error")
		log.Errorf("Error:%s", "Error")
		log.Panic("panic")
		log.Panicf("panic:%s", "panic")

		return nil
	})

	app := glue.NewApp(glue.Server(rcpSrv))
	app.Start()
}
