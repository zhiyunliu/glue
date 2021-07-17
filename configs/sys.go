package configs

import (
	"os"
	"path/filepath"
)

var PlatName string = ""
var SysName string = filepath.Base(os.Args[0])
var ClusterName string = "prod"
var IsDebug bool = false
var IPMask string = ""
var TraceType string = ""
var TracePort string = ""
var Usage string = ""
var Version string = "0.0.1"
