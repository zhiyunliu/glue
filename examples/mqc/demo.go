package main

import (
	"time"

	"github.com/zhiyunliu/velocity/context"
)

type demo struct{}

func (d *demo) Handle(ctx context.Context) interface{} {
	ctx.Log().Infof("mqc.demo", time.Now().Format("2006-01-02 15:04:05"))
	return "success"
}
