package demos

import (
	"github.com/zhiyunliu/gel"
	"github.com/zhiyunliu/gel/context"
)

type Queuedemo struct{}

func NewQueue() *Queuedemo {
	return &Queuedemo{}
}

func (d *Queuedemo) GetHandle(ctx context.Context) interface{} {
	ctx.Log().Debug("Queuedemo.get")
	queueObj := gel.Queue().GetQueue("default")

	err := queueObj.Send(ctx, "key", map[string]interface{}{
		"a": "1",
		"b": "2",
	})
	return map[string]interface{}{
		"err": err,
	}
}
