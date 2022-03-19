package velocity

import (
	"github.com/zhiyunliu/velocity/cli"
	"github.com/zhiyunliu/velocity/compatible"
	_ "github.com/zhiyunliu/velocity/encoding/json"
	_ "github.com/zhiyunliu/velocity/encoding/yaml"

	_ "github.com/zhiyunliu/velocity/contrib/registry/nacos"
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
	compatible.AppClose()
	return nil
}
