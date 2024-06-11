package log

import (
	"github.com/zhiyunliu/golibs/xlog"
	"github.com/zhiyunliu/golibs/xpath"
)

var (
	WithName    = xlog.WithName
	WithSid     = xlog.WithSid
	WithSrvType = xlog.WithSrvType
	WithField   = xlog.WithField
	WithFields  = xlog.WithFields
)

type Option = xlog.Option

type ServerOption func(opt *Options)

type Options struct {
	WithRequest  bool
	WithResponse bool
	Excludes     []string
	pathMatcher  *xpath.Match
}

func (opts *Options) IsExclude(path string) bool {
	if opts.pathMatcher == nil {
		return false
	}
	isMatch, _ := opts.pathMatcher.Match(path, "/")
	return isMatch
}

func WithRequest() ServerOption {
	return func(opt *Options) {
		opt.WithRequest = true
	}
}

func WithResponse() ServerOption {
	return func(opt *Options) {
		opt.WithResponse = true
	}
}

// Deprecated: use func (e *Server) Handle(path string, obj interface{}, opts ...engine.RouterOption) 中Opts代替
func Excludes(excludes ...string) ServerOption {
	return func(opt *Options) {
		opt.Excludes = excludes
		opt.pathMatcher = xpath.NewMatch(excludes, xpath.WithCache(false))
	}
}

type ConfigOption = xlog.ConfigOption

var (
	WithConfigPath  = xlog.WithConfigPath
	WithLayout      = xlog.WithLayout
	WithConcurrency = xlog.WithConcurrency
)
