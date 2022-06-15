package demos

import (
	"time"

	"github.com/zhiyunliu/glue/context"
)

type Fulldemo struct{}

func (d *Fulldemo) Handle(ctx context.Context) interface{} {
	ctx.Log().Infof("mqc.demo:%s", time.Now().Format("2006-01-02 15:04:05"))

	ctx.Log().Infof("header.a:%+v", ctx.Request().GetHeader("a"))
	time.Sleep(time.Millisecond * 200)

	ctx.Log().Infof("header.b:%+v", ctx.Request().GetHeader("b"))
	time.Sleep(time.Millisecond * 200)

	ctx.Log().Infof("header.c:%+v", ctx.Request().GetHeader("c"))

	ctx.Log().Infof("body-1:%s", ctx.Request().Body().Bytes())
	time.Sleep(time.Millisecond * 200)

	mapData := map[string]interface{}{}
	ctx.Request().Body().Scan(&mapData)
	ctx.Log().Infof("body-2:%+v", mapData)

	return "success"
}
