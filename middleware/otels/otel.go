package otels

import (
	"net/http"
	"strconv"
	"time"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/opentelemetry/metrics"
	gsemconv "github.com/zhiyunliu/glue/opentelemetry/semconv"
	"github.com/zhiyunliu/glue/standard"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	oteltrace "go.opentelemetry.io/otel/trace"

	stdmetrics "github.com/zhiyunliu/glue/metrics"
	"github.com/zhiyunliu/glue/middleware"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

const (
	ScopeName = "middleware/otels"
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

	stdInstance := standard.GetInstance(stdmetrics.TypeNode).(stdmetrics.StandardMetric)
	//todo
	metricsProvider := stdInstance.GetProvider("prometheus")
	mets := &metrics.Metrics{}
	metricsProvider.Build(mets)

	return func(handler middleware.Handler) middleware.Handler {
		return func(c context.Context) (reply interface{}) {
			serverKind := c.ServerType()
			savedCtx := c.Context()
			fullPath := c.Request().Path().FullPath()
			mets.RequestProcessing.Add(1, serverKind, fullPath)
			startTime := time.Now()

			defer func() {
				mets.RequestProcessing.Sub(1, serverKind, fullPath)
				c.ResetContext(savedCtx)
			}()
			attributes := gsemconv.RequestTraceAttrs(c)
			ctx := cfg.Propagators.Extract(savedCtx, c.Request().Header())
			opts := []oteltrace.SpanStartOption{
				oteltrace.WithAttributes(attributes...),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			}
			ctx, span := tracer.Start(ctx, fullPath, opts...)
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
			mets.RequestLatency.Record(time.Since(startTime), serverKind, fullPath, strconv.Itoa(statusCode))
			return reply
		}
	}
}
