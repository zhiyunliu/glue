package metrics

import (
	"github.com/zhiyunliu/glue/encoding"
	"github.com/zhiyunliu/glue/middleware"
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
