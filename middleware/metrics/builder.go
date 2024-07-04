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
func (xBuilder) Build(cfg *middleware.Config) middleware.Middleware {
	mCfg := &Config{}
	encoding.GetCodec(cfg.Data.Codec).Unmarshal(cfg.Data.Data, mCfg)
	return serverByConfig(mCfg)
}
