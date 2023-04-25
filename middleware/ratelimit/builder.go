package ratelimit

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
	return "ratelimit"
}
func (xBuilder) Build(cfg *middleware.Config) middleware.Middleware {
	data := cfg.Data
	limitCfg := &Config{}
	encoding.GetCodec(data.Codec).Unmarshal(data.Data, &limitCfg)
	return serverByConfig(limitCfg)
}
