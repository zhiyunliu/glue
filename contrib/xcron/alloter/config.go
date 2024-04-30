package alloter

import (
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/xcron"
)

type serverConfig struct {
	Config      xcron.Config        `json:"config" yaml:"config"`
	Middlewares []middleware.Config `json:"middlewares"  yaml:"middlewares"`
	Jobs        []*xcron.Job        `json:"jobs"  yaml:"jobs"`
}
