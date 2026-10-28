package ipc

import (
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/xipc"
)

type serverConfig struct {
	Config      xipc.Config         `json:"config" yaml:"config"`
	Middlewares []middleware.Config `json:"middlewares" yaml:"middlewares"`
	Header      engine.Header       `json:"header" yaml:"header"`
}

// Config 定义本地端点、引擎和停止等待上限，超时单位为秒。
