package demos

import (
	"time"

	"github.com/zhiyunliu/gel/context"
)

type Fulldemo struct{}

func (d *Fulldemo) Handle(ctx context.Context) interface{} {
	ctx.Log().Infof("cron.demo:%s", time.Now().Format("2006-01-02 15:04:05"))

	ctx.Log().Infof("body-1:%s", ctx.Request().Body().Bytes())
	time.Sleep(time.Millisecond * 200)

	mapData := map[string]interface{}{}
	ctx.Request().Body().Scan(&mapData)
	ctx.Log().Infof("body-2:%+v", mapData)

	return "success"
}

func (d *Fulldemo) NoneBodyHandle(ctx context.Context) interface{} {
	ctx.Log().Infof("cron.NoneBody:%s", time.Now().Format("2006-01-02 15:04:05"))

	ctx.Log().Infof("NoneBody-1:%s", ctx.Request().Body().Bytes())
	time.Sleep(time.Millisecond * 200)

	mapData := map[string]interface{}{}
	ctx.Request().Body().Scan(&mapData)
	ctx.Log().Infof("NoneBody-2:%+v", mapData)

	return "success"
}

func (d *Fulldemo) NotRunHandle(ctx context.Context) interface{} {
	ctx.Log().Infof("cron.NotRun:%s", time.Now().Format("2006-01-02 15:04:05"))

	ctx.Log().Infof("NoneBody-1:%s", ctx.Request().Body().Bytes())
	time.Sleep(time.Millisecond * 200)

	mapData := map[string]interface{}{}
	ctx.Request().Body().Scan(&mapData)
	ctx.Log().Infof("NoneBody-2:%+v", mapData)

	return "success"
}
