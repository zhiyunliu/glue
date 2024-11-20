package metrics

import (
	"fmt"

	"github.com/zhiyunliu/xbinding"

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
