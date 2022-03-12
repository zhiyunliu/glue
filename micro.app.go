package velocity

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/zhiyunliu/velocity/appcli"
	"github.com/zhiyunliu/velocity/registry"
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
	opts     *appcli.Options
	cliApp   *appcli.App
	instance *registry.ServiceInstance
}

//NewApp 创建微服务应用
func NewApp(opts ...Option) (m *MicroApp) {
	o := &appcli.Options{}
	if id, err := uuid.NewUUID(); err == nil {
		o.Id = id.String()
	}
	for _, opt := range opts {
		opt(o)
	}
	m = &MicroApp{opts: o}
	return m
}

//Start 启动服务器
func (m *MicroApp) Start() error {
	if len(m.opts.servers) == 0 {
		return fmt.Errorf("没有需要启动都服务应用")
	}
	m.cliApp = appcli.New(m.opts)
	return m.cliApp.Start()
}

//Close 关闭服务器
func (m *MicroApp) Stop() error {

	return m.cliApp.Stop()
}

// Name returns service name.
func (a *MicroApp) Name() string { return a.opts.name }

// Version returns app version.
func (a *MicroApp) Version() string { return a.opts.version }

// Metadata returns service metadata.
func (a *MicroApp) Metadata() map[string]string { return a.opts.metadata }

// Endpoint returns endpoints.
func (a *MicroApp) Endpoint() []string {
	if a.instance == nil {
		return []string{}
	}
	return a.instance.Endpoints
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
