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
