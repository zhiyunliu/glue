package otels

import (
	sctx "context"
	"net/http"
	"time"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"

	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/zhiyunliu/glue/middleware"
	gsemconv "github.com/zhiyunliu/glue/middleware/otels/semconv"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

const (
	ScopeName = "github.com/zhiyunliu/glue/middleware/otels"
)

// Server is middleware server-side metrics.
func serverByOptions(op *options) middleware.Middleware {

	cfg := config{}

	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	tracer := cfg.TracerProvider.Tracer(
		ScopeName,
		oteltrace.WithInstrumentationVersion(Version()),
	)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}
	if cfg.MeterProvider == nil {
		cfg.MeterProvider = otel.GetMeterProvider()
	}
	meter := cfg.MeterProvider.Meter(
		ScopeName,
		metric.WithInstrumentationVersion(Version()),
	)

	return func(handler middleware.Handler) middleware.Handler {

		return func(c context.Context) (reply interface{}) {

			requestStartTime := time.Now()
			fullPath := c.Request().Path().FullPath()
			savedCtx := c.Context()
			defer func() {
				c.ResetContext(savedCtx)
			}()
			additionalAttributes := gsemconv.RequestTraceAttrs(op.svcName, c.Request())
			ctx := cfg.Propagators.Extract(savedCtx, c.Request().Header())
			opts := []oteltrace.SpanStartOption{
				oteltrace.WithAttributes(additionalAttributes...),
				oteltrace.WithAttributes(semconv.HTTPRoute(fullPath)),
				oteltrace.WithAttributes(gsemconv.XRequestID(c.Request().RequestID())),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			}
			var spanName string = c.Request().Path().FullPath()
			ctx, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()

			// pass the span through the request context
			//c.Request = c.ResetContext(ctx)
			c.ResetContext(ctx)
			reply = handler(c)

			var (
				statusCode int = http.StatusOK
				err        error
			)

			if respErr, ok := reply.(errors.RespError); ok {
				statusCode = respErr.GetCode()
				if statusCode == 0 {
					statusCode = c.Response().GetStatusCode()
				}
				if err, ok = reply.(error); !ok {
					err = errors.New(statusCode, respErr.GetMessage())
				}

			} else if rerr, ok := reply.(error); ok {
				statusCode = http.StatusInternalServerError
				err = rerr
				if se := errors.FromError(rerr); se != nil {
					statusCode = int(se.Code)
				}
			}

			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			span.SetStatus(gsemconv.Status(statusCode))
			if statusCode > 0 {
				span.SetAttributes(semconv.HTTPResponseStatusCode(statusCode))
			}

			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			}

			RecordMetrics(ctx, gsemconv.ServerMetricData{
				ServerName:   op.svcName,
				ResponseSize: int64(c.Response().Size()),
				MetricAttributes: gsemconv.MetricAttributes{
					Req:                  c.Request(),
					StatusCode:           statusCode,
					AdditionalAttributes: additionalAttributes,
				},
				MetricData: gsemconv.MetricData{
					RequestSize: c.Request().GetContentLength(),
					ElapsedTime: float64(time.Since(requestStartTime)) / float64(time.Millisecond),
				},
			})
			return reply
		}
	}
}

func RecordMetrics(ctx sctx.Context, md gsemconv.ServerMetricData) {

	if s.requestBytesCounter != nil && s.responseBytesCounter != nil && s.serverLatencyMeasure != nil {
		attributes := OldHTTPServer{}.MetricAttributes(md.ServerName, md.Req, md.StatusCode, md.AdditionalAttributes)
		o := metric.WithAttributeSet(attribute.NewSet(attributes...))
		addOpts := metricAddOptionPool.Get().(*[]metric.AddOption)
		*addOpts = append(*addOpts, o)
		s.requestBytesCounter.Add(ctx, md.RequestSize, *addOpts...)
		s.responseBytesCounter.Add(ctx, md.ResponseSize, *addOpts...)
		s.serverLatencyMeasure.Record(ctx, md.ElapsedTime, o)
		*addOpts = (*addOpts)[:0]
		metricAddOptionPool.Put(addOpts)
	}

	if s.duplicate && s.requestDurationHistogram != nil && s.requestBodySizeHistogram != nil && s.responseBodySizeHistogram != nil {
		attributes := CurrentHTTPServer{}.MetricAttributes(md.ServerName, md.Req, md.StatusCode, md.AdditionalAttributes)
		o := metric.WithAttributeSet(attribute.NewSet(attributes...))
		recordOpts := metricRecordOptionPool.Get().(*[]metric.RecordOption)
		*recordOpts = append(*recordOpts, o)
		s.requestBodySizeHistogram.Record(ctx, md.RequestSize, *recordOpts...)
		s.responseBodySizeHistogram.Record(ctx, md.ResponseSize, *recordOpts...)
		s.requestDurationHistogram.Record(ctx, md.ElapsedTime/1000.0, o)
		*recordOpts = (*recordOpts)[:0]
		metricRecordOptionPool.Put(recordOpts)
	}

}
