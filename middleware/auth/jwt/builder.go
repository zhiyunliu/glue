package jwt

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
	return "jwt"
}
func (xBuilder) Build(cfg *middleware.Config) (middleware.Middleware, error) {
	data := cfg.Data
	authCfg := &Config{}

	codec, err := xbinding.GetCodec(xbinding.WithContentType(data.Codec))
	if err != nil {
		return nil, fmt.Errorf("jwt err:%w", err)
	}

	if err = codec.Bind(xbinding.BytesReader(data.Data), authCfg); err != nil {
		return nil, err
	}

	return serverByConfig(authCfg), nil
}
