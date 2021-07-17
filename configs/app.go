package configs

type AppSetting struct {
	Addr        string
	PlatName    string
	SysName     string
	ClusterName string
	IsDebug     bool
	IPMask      string
	TraceType   string
	TracePort   string
	Usage       string
	Version     string
}
