package gel

import (
	"runtime"

	"github.com/zhiyunliu/gel/cli"
	"github.com/zhiyunliu/gel/compatible"
	_ "github.com/zhiyunliu/gel/encoding/json"
	_ "github.com/zhiyunliu/gel/encoding/yaml"
	"github.com/zhiyunliu/gel/log"
)

//MicroApp  微服务应用
type MicroApp struct {
	opts   []Option
	cliApp *cli.App
}

//NewApp 创建微服务应用
func NewApp(opts ...Option) (m *MicroApp) {
	m = &MicroApp{opts: opts}
	m.cliApp = cli.New(opts...)
	log.Concurrency(runtime.NumCPU() * 20)
	return m
}

//Start 启动服务器
func (m *MicroApp) Start() error {

	return m.cliApp.Start()
}

//Close 关闭服务器
func (m *MicroApp) Stop() error {
	return compatible.AppClose()
}
