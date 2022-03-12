package velocity

import (
	"net/url"

	"github.com/zhiyunliu/velocity/appcli"
	"github.com/zhiyunliu/velocity/log"
	"github.com/zhiyunliu/velocity/transport"
)

// Option is an application option.
type Option func(o *appcli.Options)

// // options is an application options.
// type options struct {
// 	id        string
// 	name      string
// 	version   string
// 	metadata  map[string]string
// 	endpoints []*url.URL

// 	logger           *log.Wrapper
// 	registrar        registry.Registrar
// 	registrarTimeout time.Duration
// 	stopTimeout      time.Duration
// 	servers          []transport.Server
// }

// ID with service id.
func ID(id string) Option {
	return func(o *appcli.Options) { o.Id = id }
}

// Name with service name.
func Name(name string) Option {
	return func(o *appcli.Options) { o.Name = name }
}

// Version with service version.
func Version(version string) Option {
	return func(o *appcli.Options) { o.Version = version }
}

// Metadata with service metadata.
func Metadata(md map[string]string) Option {
	return func(o *appcli.Options) { o.Metadata = md }
}

// Endpoint with service endpoint.
func Endpoint(endpoints ...*url.URL) Option {
	return func(o *appcli.Options) { o.Endpoints = endpoints }
}

// Logger with service logger.
func Logger(logger log.Logger) Option {
	return func(o *appcli.Options) {
		o.Logger = log.NewWrapper(logger)
	}
}

// Server with transport servers.
func Server(srv ...transport.Server) Option {
	return func(o *appcli.Options) { o.Servers = srv }
}
