package global

import (
	"fmt"
	"runtime"
)

var (
	GitCommit   = "unknown"
	BuildTime   = "unknown"
	Version     = "unknown"
	PkgVersion  = "unknown"
	DisplayName = ""
	Usage       = "unknown"
)

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
	`,
		GitCommit,
		BuildTime,
		Version,
		PkgVersion,
		DisplayName,
		runtime.Version(),
		Usage,
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
