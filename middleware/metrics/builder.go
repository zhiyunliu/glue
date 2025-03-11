package metrics

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
	return "metrics"
}
func (xBuilder) Build(cfg *middleware.Config) (middleware.Middleware, error) {
	mCfg := &Config{}
	data := cfg.Data

	codec, err := xbinding.GetCodec(xbinding.WithContentType(data.Codec))
	if err != nil {
		return nil, fmt.Errorf("metrics err:%w", err)
	}

	if err = codec.Bind(xbinding.BytesReader(data.Data), mCfg); err != nil {
		return nil, err
	}

	return serverByConfig(mCfg), nil
}
