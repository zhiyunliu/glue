// Package compatible darwin (mac os x) version
package compatible

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

//CheckPrivileges 检查是否有管理员权限
func CheckPrivileges() error {
	if output, err := exec.Command("id", "-g").Output(); err == nil {
		if gid, parseErr := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 32); parseErr == nil {
			if gid == 0 {
				return nil
			}
			return errRootPrivileges
		}
	}
	return errUnsupportedSystem
}

var CmdsRunNotifySignals = []os.Signal{os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR2}

//CmdsUpdateProcessSignal CmdsUpdateProcessSignal
var CmdsUpdateProcessSignal = syscall.SIGUSR2

//AppClose AppClose
func AppClose() error {
	parent := syscall.Getpid()
	return syscall.Kill(parent, syscall.SIGTERM)
}
