package otels

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/metrics"
	"github.com/zhiyunliu/glue/opentelemetry"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"

	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/zhiyunliu/glue/middleware"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

const (
	ScopeName = "glue-otels"
	maxNumber = 1000
)

// Server is middleware server-side metrics.
func Server() middleware.Middleware {

	cfg := config{}
	if cfg.TracerProvider == nil {
		cfg.TracerProvider = otel.GetTracerProvider()
	}
	tracer := cfg.TracerProvider.Tracer(ScopeName)
	if cfg.Propagators == nil {
		cfg.Propagators = otel.GetTextMapPropagator()
	}
	if cfg.MeterProvider == nil {
		cfg.MeterProvider = otel.GetMeterProvider()
	}
	mets := &Metrics{}
	factory := metrics.NewFactory(cfg.MeterProvider, ScopeName)

	err := metrics.Init(mets, factory)
	if err != nil {
		panic(fmt.Errorf("metrics: %s", err.Error()))
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(c context.Context) (reply interface{}) {
			serverKind := c.ServerType()
			savedCtx := c.Context()
			fullPath := c.Request().Path().FullPath()

			metricAttrs := []attribute.KeyValue{
				attribute.String("kind", serverKind),
				attribute.String("path", fullPath),
			}

			mets.RequestProcessing.Record(savedCtx, 1, metric.WithAttributes(metricAttrs...))
			startTime := time.Now()

			defer func() {
				mets.RequestProcessing.Record(savedCtx, -1, metric.WithAttributes(metricAttrs...))
				c.ResetContext(savedCtx)
			}()
			attributes := requestTraceAttrs(c)
			ctx := cfg.Propagators.Extract(savedCtx, c.Request().Header())
			opts := []oteltrace.SpanStartOption{
				oteltrace.WithAttributes(attributes...),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			}
			ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s", serverKind, fullPath), opts...)
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
			span.SetStatus(opentelemetry.Status(statusCode))
			if statusCode > 0 {
				span.SetAttributes(semconv.HTTPResponseStatusCode(statusCode))
			}

			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			}
			mets.RequestCounter.Add(savedCtx, 1, metric.WithAttributes(
				attribute.String("kind", serverKind),
				attribute.String("path", fullPath),
				attribute.Int("code", statusCode),
			))

			mets.RequestLatency.Record(ctx, startTime, metric.WithAttributes(metricAttrs...))
			return reply
		}
	}
}

func requestTraceAttrs(ctx context.Context) []attribute.KeyValue {
	count := 7 // ServerAddress, Method, Scheme
	req := ctx.Request()
	clientIP := req.GetClientIP()

	attrs := make([]attribute.KeyValue, 0, count)
	attrs = append(attrs,
		attribute.String("content.type", req.ContentType()),
		attribute.String("server.type", ctx.ServerType()),
		semconv.HTTPRequestMethodKey.String(req.GetMethod()),
		opentelemetry.XRequestID(req.RequestID()),
	)
	query := req.Query()
	if qval := query.GetValues(); len(qval) > 0 {
		attrs = append(attrs, semconv.URLQuery(qval.Encode()))
	}

	if useragent := req.GetHeader("User-Agent"); useragent != "" {
		attrs = append(attrs, semconv.UserAgentName(useragent))
	}

	if clientIP != "" {
		attrs = append(attrs, semconv.ClientAddress(clientIP))
	}

	return attrs
}
