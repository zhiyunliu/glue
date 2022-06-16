package cli

import (
	"net/url"
	"time"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/registry"
	"github.com/zhiyunliu/glue/transport"
)

type Options struct {
	Id        string
	Metadata  map[string]string
	Endpoints []*url.URL

	Registrar        registry.Registrar
	Config           config.Config
	RegistrarTimeout time.Duration
	StopTimeout      time.Duration
	Servers          []transport.Server

	setting  *appSetting
	initFile string
}

//Option 配置选项
type Option func(*Options)

// ID with service id.
func ID(id string) Option {
	return func(o *Options) { o.Id = id }
}

// Metadata with service metadata.
func Metadata(md map[string]string) Option {
	return func(o *Options) { o.Metadata = md }
}

// Endpoint with service endpoint.
func Endpoint(endpoints ...*url.URL) Option {
	return func(o *Options) { o.Endpoints = endpoints }
}

// Server with transport servers.
func Server(srv ...transport.Server) Option {
	return func(o *Options) { o.Servers = srv }
}

//WithAppMode
func WithAppMode(mode AppMode) Option {
	return func(o *Options) {
		o.setting.Mode = mode
	}
}

//IpMask
func IpMask(mask string) Option {
	return func(o *Options) {
		o.setting.IpMask = mask
	}
}

//TraceAddr
func TraceAddr(addr string) Option {
	return func(o *Options) {
		o.setting.TraceAddr = addr
	}
}

// ServiceOption
func ServiceOption(key string, val interface{}) Option {
	return func(o *Options) {
		o.setting.Options[key] = val
	}
}

// ServiceDependencies
func ServiceDependencies(dependencies ...string) Option {
	return func(o *Options) {
		o.setting.Dependencies = dependencies
	}
}

// ServiceDependencies
func LogConcurrency(concurrency int) Option {
	return func(o *Options) {
		log.Concurrency(concurrency)
	}
}

type AppMode string

const (
	Debug   AppMode = "debug"
	Release AppMode = "release"
)

type appSetting struct {
	Mode         AppMode                `json:"mode"`
	TraceAddr    string                 `json:"trace_addr"`
	IpMask       string                 `json:"ip_mask"`
	Dependencies []string               `json:"dependencies"`
	Options      map[string]interface{} `json:"options"`
}
