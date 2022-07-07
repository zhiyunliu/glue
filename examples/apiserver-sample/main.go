package main

import (
	"github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/server/api"
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

	apiSrv := api.New("apiserver", api.WithServiceName("demo"))
	apiSrv.Handle("/demo/origin", func(ctx context.Context) interface{} {
		ver := ctx.Request().Query().Get("ver")
		b := &body{}
		ctx.Bind(b)
		return map[string]interface{}{
			"v": ver,
			"b": b.Seq,
		}
	})
	apiSrv.Handle("/demo/struct", &demo{})
	apiSrv.Handle("/demo", func(ctx context.Context) interface{} {
		return map[string]interface{}{
			"a": "1",
		}
	})
	app := glue.NewApp(glue.Server(apiSrv))
	app.Start()
}
