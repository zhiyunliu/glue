package jwt

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
	return "jwt"
}
func (xBuilder) Build(cfg *middleware.Config) middleware.Middleware {
	data := cfg.Data
	authCfg := &Config{}
	encoding.GetCodec(data.Codec).Unmarshal(data.Data, &authCfg)

	return serverByConfig(authCfg)
}
