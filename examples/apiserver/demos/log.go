package demos

import (
	"strconv"

	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/golibs/xlog"
)

type Logdemo struct{}

func NewLogDemo() *Logdemo {
	return &Logdemo{}
}

func (d *Logdemo) ConcurrencyHandle(ctx context.Context) interface{} {
	cntVal := ctx.Request().Query().Get("cnt")
	cnt, _ := strconv.Atoi(cntVal)
	xlog.Concurrency(cnt)
	return map[string]interface{}{
		"status": "success",
	}

}

func (d *Logdemo) InfoHandle(ctx context.Context) interface{} {
	return xlog.Stats()
}
