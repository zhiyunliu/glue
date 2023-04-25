package tracing

import (
	"github.com/zhiyunliu/glue/encoding"
	"github.com/zhiyunliu/glue/middleware"
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
func (xBuilder) Build(cfg *middleware.Config) middleware.Middleware {
	data := cfg.Data
	traceCfg := &Config{}
	encoding.GetCodec(data.Codec).Unmarshal(data.Data, &traceCfg)
	switch traceCfg.SpanKind {
	case trace.SpanKindClient:
		fallthrough
	case trace.SpanKindProducer:
		return clientByConfig(traceCfg)
	case trace.SpanKindServer:
		fallthrough
	case trace.SpanKindConsumer:
		return serverByConfig(traceCfg)
	default:
		return serverByConfig(traceCfg)
	}
}
