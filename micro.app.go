package velocity

import (
	"context"

	"github.com/zhiyunliu/velocity/appcli"
	"github.com/zhiyunliu/velocity/compatible"
)

type AppInfo interface {
	ID() string
	Name() string
	Version() string
	Metadata() map[string]string
	Endpoint() []string
}

//MicroApp  微服务应用
type MicroApp struct {
	opts   []Option
	cliApp *appcli.App
}

//NewApp 创建微服务应用
func NewApp(opts ...Option) (m *MicroApp) {
	m = &MicroApp{opts: opts}
	m.cliApp = appcli.New(opts...)
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

type appKey struct{}

// NewContext returns a new Context that carries value.
func NewContext(ctx context.Context, s AppInfo) context.Context {
	return context.WithValue(ctx, appKey{}, s)
}

// FromContext returns the Transport value stored in ctx, if any.
func FromContext(ctx context.Context) (s AppInfo, ok bool) {
	s, ok = ctx.Value(appKey{}).(AppInfo)
	return
}
