package alloter

import (
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/xmqc"
)

type serverConfig struct {
	Config      xmqc.Config         `json:"config" yaml:"config"`
	Middlewares []middleware.Config `json:"middlewares"  yaml:"middlewares"`
	Tasks       xmqc.TaskList       `json:"tasks"  yaml:"tasks"`
}
