package http

import (
	"net/http"
	"time"

	"github.com/zhiyunliu/glue/constants"

	"github.com/zhiyunliu/glue/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var _ http.RoundTripper = &TracerTransport{}

type TracerTransport struct {
	base http.RoundTripper
}

func NewTransport(base http.RoundTripper) *TracerTransport {
	if base == nil {
		base = http.DefaultTransport
	}

	t := TracerTransport{
		base: base,
	}
	return &t
}

func (t *TracerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	requestStartTime := time.Now()

	var tp trace.TracerProvider
	if span := trace.SpanFromContext(req.Context()); span.SpanContext().IsValid() {
		tp = span.TracerProvider()
	} else {
		tp = otel.GetTracerProvider()
	}

	tracer := tp.Tracer("http-client")
	// 创建span
	ctx, span := tracer.Start(req.Context(),
		"HTTP Client "+req.Method,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("http.method", req.Method),
			attribute.String("http.url", req.URL.String()),
			attribute.Int64("http.request.content_length", req.ContentLength),
		),
	)
	defer span.End()

	// 注入跟踪信息到请求头
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	if xrequestId := req.Header.Get(constants.HeaderRequestId); xrequestId != "" {
		span.SetAttributes(opentelemetry.XRequestID(xrequestId)) // 添加X-Request-ID到span的属性
	}

	res, err := t.base.RoundTrip(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return res, err
	}

	elapsedTime := float64(time.Since(requestStartTime)) / float64(time.Millisecond)
	span.SetAttributes(
		attribute.Int("http.response.status_code", res.StatusCode),
		attribute.Float64("http.response.elapsed_time", elapsedTime),
	)
	if res.StatusCode >= http.StatusBadRequest {
		span.SetStatus(codes.Error, http.StatusText(res.StatusCode))
	}
	return res, err
}
