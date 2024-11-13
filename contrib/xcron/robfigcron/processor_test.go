package robfigcron

import (
	"context"
	"net/http"
	"testing"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/zhiyunliu/alloter"
	"github.com/zhiyunliu/glue/xcron"
	"github.com/zhiyunliu/golibs/xlist"
)

func Test_processor_handleImmediatelyJob(t *testing.T) {
	processor := &processor{
		ctx:             context.Background(),
		closeChan:       make(chan struct{}),
		jobs:            cmap.New(),
		monopolyJobs:    cmap.New(),
		routerEngine:    alloter.New(),
		cronStdEngine:   cron.New(),
		cronSecEngine:   cron.New(cron.WithSeconds()),
		immediatelyJobs: xlist.NewList(),
	}

	expectResult1 := []int{1}
	expectResult := []int{1, 2}
	goResult := []int{}

	processor.routerEngine.Handle(http.MethodPost, "/test/1", func(ctx *alloter.Context) {
		goResult = append(goResult, 1)
	})

	processor.routerEngine.Handle(http.MethodPost, "/test/2", func(ctx *alloter.Context) {
		goResult = append(goResult, 2)
	})
	processor.routerEngine.Handle(http.MethodPost, "/test/3", func(ctx *alloter.Context) {
		goResult = append(goResult, 3)
	})
	processor.routerEngine.Handle(http.MethodPost, "/test/4", func(ctx *alloter.Context) {
		goResult = append(goResult, 4)
	})

	//1. 处理配置加载
	processor.Add(&xcron.Job{
		Cron:        "@every 60s",
		Service:     "/test/1",
		Disable:     false,
		Immediately: true,
		Monopoly:    false,
		WithSeconds: false},
	)

	//1. 处理配置加载
	processor.Add(&xcron.Job{
		Cron:        "@every 60s",
		Service:     "/test/3",
		Disable:     false,
		Immediately: false,
		Monopoly:    false,
		WithSeconds: true},
	)

	//启动应用
	go processor.handleImmediatelyJob()

	time.Sleep(time.Second * 2)

	assert.Equal(t, expectResult1, goResult)

	//1. 处理配置加载
	processor.Add(&xcron.Job{
		Cron:        "@every 1h",
		Service:     "/test/2",
		Disable:     false,
		Immediately: true,
		Monopoly:    false,
		WithSeconds: false},
	)
	processor.Add(&xcron.Job{
		Cron:        "@every 1h",
		Service:     "/test/4",
		Disable:     false,
		Immediately: false,
		Monopoly:    false,
		WithSeconds: true},
	)

	time.Sleep(time.Second * 4)

	assert.Equal(t, expectResult, goResult)
}
