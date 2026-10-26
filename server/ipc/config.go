package ipc

import (
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
)

type serverConfig struct {
	Config      Config              `json:"config" yaml:"config"`
	Middlewares []middleware.Config `json:"middlewares" yaml:"middlewares"`
	Header      engine.Header       `json:"header" yaml:"header"`
}

// Config 定义本地端点、引擎和停止等待上限，超时单位为秒。
type Config struct {
	Addr               string        `json:"addr" yaml:"addr"`
	Engine             string        `json:"engine" yaml:"engine"`
	Status             engine.Status `json:"status" yaml:"status"`
	StopMaximumTimeout uint          `json:"stop_maximum_timeout" yaml:"stop_maximum_timeout"`
}
