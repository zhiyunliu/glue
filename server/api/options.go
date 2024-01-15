package api

import (
	"net/http"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/log"
)

// Option 参数设置类型
type Option func(*options)

type options struct {
	serviceName string
	srvCfg      *serverConfig
	logOpts     *log.Options
	config      config.Config
	handler     http.Handler
	router      *engine.RouterGroup
	static      map[string]Static
	decReq      engine.DecodeRequestFunc
	encResp     engine.EncodeResponseFunc
	encErr      engine.EncodeErrorFunc

	startedHooks []engine.Hook
	endHooks     []engine.Hook
}

func setDefaultOption() *options {
	return &options{
		srvCfg: &serverConfig{
			Config: Config{
				Addr:              ":8080",
				Engine:            "gin",
				Status:            engine.StatusStart,
				ReadTimeout:       15,
				WriteTimeout:      15,
				ReadHeaderTimeout: 15,
				MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
			},
		},
		logOpts:      &log.Options{},
		static:       make(map[string]Static),
		startedHooks: make([]engine.Hook, 0),
		endHooks:     make([]engine.Hook, 0),
		decReq:       engine.DefaultRequestDecoder,
		encResp:      engine.DefaultResponseEncoder,
		encErr:       engine.DefaultErrorEncoder,
		router:       engine.NewRouterGroup(""),
	}
}

// WithAddr 设置服务地址
func WithAddr(addr string) Option {
	return func(o *options) {
		o.srvCfg.Config.Addr = addr
	}
}

// WithEngine 设置服务引擎（gin)
func WithEngine(engine string) Option {
	return func(o *options) {
		o.srvCfg.Config.Engine = engine
	}
}

// WithReadTimeout 读取超时时间
func WithReadTimeout(readTimeout uint) Option {
	return func(o *options) {
		o.srvCfg.Config.ReadTimeout = readTimeout
	}
}

// WithWriteTimeout 写入超时时间
func WithWriteTimeout(writeTimeout uint) Option {
	return func(o *options) {
		o.srvCfg.Config.WriteTimeout = writeTimeout
	}
}

// ReadHeaderTimeout 写入超时时间
func WithReadHeaderTimeout(readHeaderTimeout uint) Option {
	return func(o *options) {
		o.srvCfg.Config.ReadHeaderTimeout = readHeaderTimeout
	}
}

// WithMaxHeaderBytes 最大头部大小（http.DefaultMaxHeaderBytes untyped int = 1 << 20 )
func WithMaxHeaderBytes(maxHeaderBytes uint) Option {
	return func(o *options) {
		o.srvCfg.Config.MaxHeaderBytes = maxHeaderBytes
	}
}

func WithEndHook(f engine.Hook) Option {
	return func(o *options) {
		o.endHooks = append(o.endHooks, f)
	}
}

// WithStartedHook 设置启动回调函数
func WithStartedHook(f engine.Hook) Option {
	return func(o *options) {
		o.startedHooks = append(o.startedHooks, f)
	}
}

// WithStartedHook 设置启动回调函数
func WithConfig(config config.Config) Option {
	return func(o *options) {
		o.config = config
	}
}

// WithServiceName 设置服务名称
func WithServiceName(serviceName string) Option {
	return func(o *options) {
		o.serviceName = serviceName
	}
}

// Log 设置日志配置
func Log(opts ...log.ServerOption) Option {
	return func(o *options) {
		for i := range opts {
			opts[i](o.logOpts)
		}
	}
}

// WithDecodeRequestFunc 解析入参
func WithDecodeRequestFunc(decReq engine.DecodeRequestFunc) Option {
	return func(o *options) {
		o.decReq = decReq
	}
}

// WithEncodeResponseFunc 编码响应
func WithEncodeResponseFunc(encResp engine.EncodeResponseFunc) Option {
	return func(o *options) {
		o.encResp = encResp
	}
}

// WithEncodeErrorFunc 编码错误
func WithEncodeErrorFunc(encErr engine.EncodeErrorFunc) Option {
	return func(o *options) {
		o.encErr = encErr
	}
}
