package velocity

import "github.com/zhiyunliu/velocity/cli"

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
var Metadata = cli.Metadata
var Endpoint = cli.Endpoint
var Server = cli.Server
