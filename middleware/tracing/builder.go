package tracing

import (
	"fmt"

	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/xbinding"
	"go.opentelemetry.io/otel/trace"
)

func NewBuilder() middleware.MiddlewareBuilder {
	return &xBuilder{}
}

type xBuilder struct {
}

func (xBuilder) Name() string {
	return "tracing"
}
func (xBuilder) Build(cfg *middleware.Config) (middleware.Middleware, error) {
	data := cfg.Data
	traceCfg := &Config{}

	codec, err := xbinding.GetCodec(xbinding.WithContentType(data.Codec))
	if err != nil {
		return nil, fmt.Errorf("tracing err:%w", err)
	}

	if err = codec.Bind(xbinding.BytesReader(data.Data), traceCfg); err != nil {
		return nil, err
	}

	switch traceCfg.SpanKind {
	case trace.SpanKindClient:
		fallthrough
	case trace.SpanKindProducer:
		return clientByConfig(traceCfg), nil
	case trace.SpanKindServer:
		fallthrough
	case trace.SpanKindConsumer:
		return serverByConfig(traceCfg), nil
	default:
		return serverByConfig(traceCfg), nil
	}
}
