package appcli

import (
	"net/url"
	"time"

	"github.com/zhiyunliu/velocity/log"
	"github.com/zhiyunliu/velocity/registry"
	"github.com/zhiyunliu/velocity/transport"
)

type cliOptions struct {
	IsDebug  bool
	IPMask   string
	File     string
	Registry string
}

type Options struct {
	Id        string
	Name      string
	Version   string
	Metadata  map[string]string
	Endpoints []*url.URL

	Logger           *log.Wrapper
	Registrar        registry.Registrar
	RegistrarTimeout time.Duration
	StopTimeout      time.Duration
	Servers          []transport.Server
}

//Option 配置选项
type Option func(*Options)

// ID with service id.
func ID(id string) Option {
	return func(o *Options) { o.Id = id }
}

// Name with service name.
func Name(name string) Option {
	return func(o *Options) { o.Name = name }
}

// Version with service version.
func Version(version string) Option {
	return func(o *Options) { o.Version = version }
}

// Metadata with service metadata.
func Metadata(md map[string]string) Option {
	return func(o *Options) { o.Metadata = md }
}

// Endpoint with service endpoint.
func Endpoint(endpoints ...*url.URL) Option {
	return func(o *Options) { o.Endpoints = endpoints }
}

// Logger with service logger.
func Logger(logger log.Logger) Option {
	return func(o *Options) {
		o.Logger = log.NewWrapper(logger)
	}
}

// Server with transport servers.
func Server(srv ...transport.Server) Option {
	return func(o *Options) { o.Servers = srv }
}
