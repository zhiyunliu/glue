package defaultipc

import (
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/xipc"
)

type serverConfig struct {
	Config      *xipc.Config        `json:"config"`
	Middlewares []middleware.Config `json:"middlewares"  yaml:"middlewares"`
}
