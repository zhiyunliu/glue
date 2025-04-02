package glue

import "github.com/zhiyunliu/glue/cli"

// Option is an application option.
type Option = cli.Option

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

var ID = cli.ID

// var Name = cli.Name
// var Version = cli.Version
var (
	Metadata            = cli.Metadata
	Server              = cli.Server
	AppMode             = cli.WithAppMode
	WithConfigSource    = cli.WithConfigSource
	IpMask              = cli.IpMask
	TraceAddr           = cli.TraceAddr
	ServiceDependencies = cli.ServiceDependencies
	ServiceOption       = cli.ServiceOption
	LogParams           = cli.LogParams
	StartingHook        = cli.StartingHook
	StartedHook         = cli.StartedHook
	StopingHook         = cli.StopingHook
	StopedHook          = cli.StopedHook
	Command             = cli.Command
)
