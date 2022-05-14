package handles

import (
	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/golibs/xlog"
)

type Logdemo struct{}

func NewLogDemo() *Logdemo {
	return &Logdemo{}
}

func (d *Logdemo) InfoHandle(ctx context.Context) interface{} {
	return xlog.Stats()
}
