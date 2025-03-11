package ratelimit

import (
	"fmt"

	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/xbinding"
)

func NewBuilder() middleware.MiddlewareBuilder {
	return &xBuilder{}
}

type xBuilder struct {
}

func (xBuilder) Name() string {
	return "ratelimit"
}
func (xBuilder) Build(cfg *middleware.Config) (middleware.Middleware, error) {
	data := cfg.Data
	limitCfg := &Config{}

	codec, err := xbinding.GetCodec(xbinding.WithContentType(data.Codec))
	if err != nil {
		return nil, fmt.Errorf("ratelimit err:%w", err)
	}

	if err = codec.Bind(xbinding.BytesReader(data.Data), limitCfg); err != nil {
		return nil, err
	}

	return serverByConfig(limitCfg), nil
}
