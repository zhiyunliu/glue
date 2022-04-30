package metrics

import (
	"github.com/zhiyunliu/gel/encoding"
	"github.com/zhiyunliu/gel/middleware"
)

func NewBuilder() middleware.MiddlewareBuilder {
	return &xBuilder{}
}

type xBuilder struct {
}

func (xBuilder) Name() string {
	return "metrics"
}
func (xBuilder) Build(data middleware.RawMessage) middleware.Middleware {
	cfg := &Config{}
	encoding.GetCodec(data.Codec).Unmarshal(data.Data, &cfg)
	return serverByConfig(cfg)
}
