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
func (xBuilder) Build(data middleware.RawMessage) middleware.Middleware {
	cfg := &Config{}
	encoding.GetCodec(data.Codec).Unmarshal(data.Data, &cfg)
	switch cfg.SpanKind {
	case trace.SpanKindClient:
		fallthrough
	case trace.SpanKindProducer:
		return clientByConfig(cfg)
	case trace.SpanKindServer:
		fallthrough
	case trace.SpanKindConsumer:
		return serverByConfig(cfg)
	default:
		return serverByConfig(cfg)
	}
}
