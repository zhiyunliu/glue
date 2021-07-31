package appcli

type Options struct {
	version string
	usage   string

	Addr        string
	PlatName    string
	SysName     string
	IsDebug     bool
	IPMask      string
	TraceType   string
	TracePort   string
	Usage       string
	Version     string
}

//Option 配置选项
type Option func(*Options)

//WithVersion 设置版本号
func WithVersion(version string) Option {
	return func(o *Options) {
		o.version = version
	}
}

//WithUsage 设置使用说明
func WithUsage(usage string) Option {
	return func(o *Options) {
		o.usage = usage
	}
}
