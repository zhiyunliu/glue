package demos

import (
	gel "github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/glue/context"
)

type Queuedemo struct{}

func NewQueue() *Queuedemo {
	return &Queuedemo{}
}

func (d *Queuedemo) GetHandle(ctx context.Context) interface{} {
	ctx.Log().Debug("Queuedemo.get")
	queueObj := gel.Queue("default")

	err := queueObj.Send(ctx.Context(), "key", map[string]interface{}{
		"a": "1",
		"b": "2",
	})
	return map[string]interface{}{
		"err": err,
	}
}
