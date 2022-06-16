package cli

import (
	"sync"

	"github.com/urfave/cli"
)

var cmds []cli.Command = make([]cli.Command, 0)
var funcs []func(cfg *Options) cli.Command = make([]func(cfg *Options) cli.Command, 0, 1)
var once sync.Once

//RegisterFunc 注册函数，用于异步加载
func RegisterFunc(f ...func(cfg *Options) cli.Command) {
	funcs = append(funcs, f...)
}

//GetCmds 获取所有命令
func GetCmds(options *Options) []cli.Command {
	once.Do(func() {
		for _, f := range funcs {
			cmds = append(cmds, f(options))
		}
	})
	return cmds
}
