package compatible

import "errors"

var errUnsupportedSystem = errors.New("Unsupported system")
var errRootPrivileges = errors.New("You must have root user privileges. Possibly using 'sudo' command should help")

var (
	SUCCESS = "\033[32m\t\t\t\t\t[OK]\033[0m"     // Show colored "OK"
	FAILED  = "\033[31m\t\t\t\t\t[FAILED]\033[0m" // Show colored "FAILED"
)
