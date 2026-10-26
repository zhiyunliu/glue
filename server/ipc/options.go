package ipc

import (
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/middleware"
)

// Option 设置服务参数。
type Option func(*options)

type options struct {
	middlewares  []middleware.Middleware
	srvCfg       *serverConfig
	config       config.Config
	serviceName  string
	router       *engine.RouterGroup
	logOpts      *log.Options
	decReq       engine.DecodeRequestFunc
	startedHooks []engine.Hook
	endHooks     []engine.Hook
}

func setDefaultOption() *options {
	return &options{srvCfg: &serverConfig{Config: Config{Engine: "gin", Status: engine.StatusStart, StopMaximumTimeout: 15}}, router: engine.NewRouterGroup(""), logOpts: &log.Options{}, decReq: engine.DefaultRequestDecoder}
}

// WithAddr 设置本地端点，不包含命名管道前缀。
func WithAddr(addr string) Option { return func(o *options) { o.srvCfg.Config.Addr = addr } }

// WithEngine 设置框架路由引擎。
func WithEngine(name string) Option { return func(o *options) { o.srvCfg.Config.Engine = name } }

// WithConfig 设置框架配置对象。
func WithConfig(cfg config.Config) Option { return func(o *options) { o.config = cfg } }

// WithServiceName 设置注册服务名称。
func WithServiceName(name string) Option { return func(o *options) { o.serviceName = name } }

// WithStopMaximumTimeout 设置停止等待上限，单位为秒。
func WithStopMaximumTimeout(seconds uint) Option {
	return func(o *options) { o.srvCfg.Config.StopMaximumTimeout = seconds }
}

// WithStartedHook 设置启动成功后的回调。
func WithStartedHook(hook engine.Hook) Option {
	return func(o *options) { o.startedHooks = append(o.startedHooks, hook) }
}

// WithEndHook 设置停止完成后的回调。
func WithEndHook(hook engine.Hook) Option {
	return func(o *options) { o.endHooks = append(o.endHooks, hook) }
}

// WithDecodeRequestFunc 设置框架请求解码器。
func WithDecodeRequestFunc(fn engine.DecodeRequestFunc) Option {
	return func(o *options) { o.decReq = fn }
}

// Log 设置框架请求日志参数。
func Log(opts ...log.ServerOption) Option {
	return func(o *options) {
		for _, opt := range opts {
			opt(o.logOpts)
		}
	}
}
