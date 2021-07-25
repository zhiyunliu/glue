package velocity

import "github.com/zhiyunliu/velocity/globals"

type Option func(cfg *globals.AppSetting)

func WithPlatName(platName string) Option {
	return func(cfg *globals.AppSetting) {
		cfg.PlatName = platName
	}
}

func WithSysName(sysName string) Option {
	return func(cfg *globals.AppSetting) {
		cfg.SysName = sysName
	}
}
func WithClusterName(name string) Option {
	return func(cfg *globals.AppSetting) {
		cfg.ClusterName = name
	}
}
func WithIsDebug(debug bool) Option {
	return func(cfg *globals.AppSetting) {
		cfg.IsDebug = debug
	}
}
func WithIPMask(mask string) Option {
	return func(cfg *globals.AppSetting) {
		cfg.IPMask = mask
	}
}
func WithTraceType(traceType string) Option {
	return func(cfg *globals.AppSetting) {
		cfg.TraceType = traceType
	}
}
func WithTracePort(tracePort string) Option {
	return func(cfg *globals.AppSetting) {
		cfg.TracePort = tracePort
	}
}
func WithAddr(addr  string) Option {
	return func(cfg *globals.AppSetting) {
		cfg.Addr = addr
	}
}


func WithUsage(usage string) Option {
	return func(cfg *globals.AppSetting) {
		cfg.Usage = usage
	}
}
func WithVersion(version string) Option {
	return func(cfg *globals.AppSetting) {
		cfg.Version = version
	}
}
