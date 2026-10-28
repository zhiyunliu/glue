package xipc

import "github.com/zhiyunliu/glue/engine"

var (
	// ServerPrefix 服务端地址前缀
	ServerPrefix = "glue"
)

// Config 定义本地端点、引擎和停止等待上限，超时单位为秒。
type Config struct {
	Proto              string        `json:"proto" yaml:"proto"`
	Addr               string        `json:"addr" yaml:"addr"`
	Capactiy           int           `json:"capactiy" yaml:"capactiy"`
	Engine             string        `json:"engine" yaml:"engine"`
	Status             engine.Status `json:"status" yaml:"status"`
	StopMaximumTimeout uint          `json:"stop_maximum_timeout" yaml:"stop_maximum_timeout"`
}
