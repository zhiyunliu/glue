package glue

import (
	"context"
	_ "net/http/pprof"

	"github.com/zhiyunliu/glue/cli"
	"github.com/zhiyunliu/glue/compatible"
	_ "github.com/zhiyunliu/glue/encoding/binding"
	_ "github.com/zhiyunliu/glue/encoding/text"
	"github.com/zhiyunliu/glue/global"

	_ "github.com/zhiyunliu/glue/contrib/engine/alloter"
	_ "github.com/zhiyunliu/glue/contrib/engine/gin"
)

// MicroApp  微服务应用
type MicroApp struct {
	opts   []Option
	cliApp *cli.App
}

// NewApp 创建微服务应用
func NewApp(opts ...Option) (m *MicroApp) {
	m = &MicroApp{opts: opts}
	m.cliApp = cli.New(opts...)
	return m
}

// Start 启动服务器
func (m *MicroApp) Start() (err error) {
	var cancel context.CancelFunc
	global.Ctx, cancel = context.WithCancel(context.Background())
	err = m.cliApp.Start()
	if cancel != nil {
		cancel()
	}
	return
}

// Close 关闭服务器
func (m *MicroApp) Stop() error {
	return compatible.AppClose()
}
