package velocity

import "github.com/zhiyunliu/velocity/configs"

type Option func(cfg *configs.AppSetting)

func WithPlatName(platName string) Option {
	return func(cfg *configs.AppSetting) {
		cfg.PlatName = platName
	}
}

func WithSysName(sysName string) Option {
	return func(cfg *configs.AppSetting) {
		cfg.SysName = sysName
	}
}
func WithClusterName(name string) Option {
	return func(cfg *configs.AppSetting) {
		cfg.ClusterName = name
	}
}
func WithIsDebug(debug bool) Option {
	return func(cfg *configs.AppSetting) {
		cfg.IsDebug = debug
	}
}
func WithIPMask(mask string) Option {
	return func(cfg *configs.AppSetting) {
		cfg.IPMask = mask
	}
}
func WithTraceType(traceType string) Option {
	return func(cfg *configs.AppSetting) {
		cfg.TraceType = traceType
	}
}
func WithTracePort(tracePort string) Option {
	return func(cfg *configs.AppSetting) {
		cfg.TracePort = tracePort
	}
}
func WithAddr(addr  string) Option {
	return func(cfg *configs.AppSetting) {
		cfg.Addr = addr
	}
}


func WithUsage(usage string) Option {
	return func(cfg *configs.AppSetting) {
		cfg.Usage = usage
	}
}
func WithVersion(version string) Option {
	return func(cfg *configs.AppSetting) {
		cfg.Version = version
	}
}
