package velocity

type Options struct {
	PlatName  string
	IsDebug   bool
	TraceType string
	TracePort string
	Usage     string
	Version   string
}

type Option func(cfg *Options)

func WithPlatName(platName string) Option {
	return func(cfg *Options) {
		cfg.PlatName = platName
	}
}

func WithIsDebug(debug bool) Option {
	return func(cfg *Options) {
		cfg.IsDebug = debug
	}
}

func WithUsage(usage string) Option {
	return func(cfg *Options) {
		cfg.Usage = usage
	}
}
func WithVersion(version string) Option {
	return func(cfg *Options) {
		cfg.Version = version
	}
}
func WithTraceType(traceType string) Option {
	return func(cfg *Options) {
		cfg.TraceType = traceType
	}
}
func WithTracePort(tracePort string) Option {
	return func(cfg *Options) {
		cfg.TracePort = tracePort
	}
}
