package cmds

import (
	"sync"

	"github.com/urfave/cli"
	"github.com/zhiyunliu/velocity/globals"
)

var cmds []cli.Command = make([]cli.Command, 0)
var funcs []func(cfg *globals.AppSetting) cli.Command = make([]func(cfg *globals.AppSetting) cli.Command, 0, 1)
var once sync.Once

//RegisterFunc 注册函数，用于异步加载
func RegisterFunc(f ...func(cfg *globals.AppSetting) cli.Command) {
	funcs = append(funcs, f...)
}

//GetCmds 获取所有命令
func GetCmds(cfg *globals.AppSetting) []cli.Command {
	once.Do(func() {
		for _, f := range funcs {
			cmds = append(cmds, f(cfg))
		}
	})
	return cmds
}
