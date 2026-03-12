package global

import (
	"fmt"
	"runtime"
)

var (
	//开启无路由时候显示详细
	EnableNoRouteDetail bool = false
	Nostd               bool = false
)

var (
	GitCommit   = "unknown"
	BuildTime   = "unknown"
	Version     = "unknown"
	PkgVersion  = "unknown"
	DisplayName = ""
	Usage       = "unknown"
)

func GetGlueVersion() string {
	glueVersion, _ := GetPackageVersion("github.com/zhiyunliu/glue")
	return glueVersion
}

func GetGolibsVersion() string {
	golibsVersion, _ := GetPackageVersion("github.com/zhiyunliu/golibs")
	return golibsVersion
}

// BuildInfo returns a string containing build information.
func BuildInfo() string {

	return fmt.Sprintf(`
	GitCommit    = %s
	BuildTime    = %s
	Version      = %s
	PkgVersion   = %s	
	DisplayName  = %s
	GoVersion    = %s
	Usage        = %s
	GlueVersion  = %s
	GolibsVersion= %s
`,
		GitCommit,
		BuildTime,
		Version,
		PkgVersion,
		DisplayName,
		runtime.Version(),
		Usage,
		GetGlueVersion(),
		GetGolibsVersion(),
	)
}

var (
	running bool
)

func IsRunning() bool {
	return running
}

func StartRunning() {
	running = true
}
