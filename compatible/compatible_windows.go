// Package compatible windows version
package compatible

import (
	"os"
	"syscall"

	"github.com/zhiyunliu/velocity/log"
)

//CheckPrivileges 检查是否有管理员权限
func CheckPrivileges() error {
	return nil
}

var CmdsRunNotifySignals = []os.Signal{os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT}

//CmdsUpdateProcessSignal CmdsUpdateProcessSignal
var CmdsUpdateProcessSignal = syscall.SIGINT

//AppClose AppClose
func AppClose() error {
	pid := syscall.Getpid()

	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		log.Errorf("syscall.LoadDLL(kernel32.dll),Error:%+v", err)
		return err
	}

	p, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		log.Errorf("dll.FindProc(GenerateConsoleCtrlEvent),Error:%+v", err)
		return err
	}

	r, _, err := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if r == 0 {
		log.Errorf("prod.Call(CTRL_BREAK_EVENT,pid=%d),Error:%+v", pid, err)
		return err
	}
	return nil
}

func init() {
	SUCCESS = "\t\t\t\t\t[OK]"    // Show colored "OK"
	FAILED = "\t\t\t\t\t[FAILED]" // Show colored "FAILED"
}
