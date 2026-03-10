package global

var (
	//开启无路由时候显示详细
	EnableNoRouteDetail bool = false
)

var (
	GitCommit   = "unknown"
	BuildTime   = "unknown"
	Version     = "unknown"
	PkgVersion  = "unknown"
	DisplayName = ""
	Usage       = "unknown"
)

var (
	running bool
)

func IsRunning() bool {
	return running
}

func StartRunning() {
	running = true
}
