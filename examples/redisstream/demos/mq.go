package demos

import (
	"fmt"
	"time"

	"github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/glue/context"
)

func NewMQ() *MQ {
	return &MQ{}
}

type MQ struct{}

func (q *MQ) EnHandle(ctx context.Context) interface{} {

	qobj := glue.Queue("streamredis")
	err := qobj.Send(ctx.Context(), "queue1", map[string]interface{}{
		"a": "1",
		"b": 2,
		"t": time.Now().Unix(),
	})
	if err != nil {
		ctx.Log().Errorf("queue1:%+v", err)
		return err
	}
	return "success"
}

type MQC struct{}

func NewMQC() *MQC {
	return &MQC{}
}

type DemoQueueData struct {
	A string `json:"a"`
	B int    `json:"b"`
}

func (q *MQC) Handle(ctx context.Context) interface{} {
	data := &DemoQueueData{}
	ctx.Log().Debugf("header:%+v", ctx.Request().Header().Values())
	ctx.Log().Infof("body:%s", ctx.Request().Body().Bytes())
	if err := ctx.Bind(data); err != nil {
		return err
	}
	ctx.Log().Infof("data:%+v", data)
	if time.Now().Unix()%2 == 0 {
		return nil
	}
	return fmt.Errorf("error")
}
