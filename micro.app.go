package glue

import (
	_ "net/http/pprof"

	"github.com/zhiyunliu/glue/cli"
	"github.com/zhiyunliu/glue/compatible"
	_ "github.com/zhiyunliu/glue/encoding/json"
	_ "github.com/zhiyunliu/glue/encoding/yaml"
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
