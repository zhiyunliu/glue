package handles

import (
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/golibs/xlog"
)

type Logdemo struct{}

func NewLogDemo() *Logdemo {
	return &Logdemo{}
}

func (d *Logdemo) InfoHandle(ctx context.Context) interface{} {

	ctx.Log().Info(ctx.Request().Header())
	ctx.Log().Info(string(ctx.Request().Body().Bytes()))

	return xlog.Stats()
}
