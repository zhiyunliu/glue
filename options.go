package velocity

import (
	"github.com/zhiyunliu/velocity/appcli"
)

// Option is an application option.
type Option = appcli.Option

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

var ID = appcli.ID
var Name = appcli.Name
var Version = appcli.Version
var Metadata = appcli.Metadata
var Endpoint = appcli.Endpoint
var Logger = appcli.Logger
var Server = appcli.Server
