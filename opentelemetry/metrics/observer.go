// Copyright (c) 2023 The Jaeger Authors.
// Copyright (c) 2017 Uber Technologies, Inc.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/zhiyunliu/glue/metrics"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

const defaultMaxNumberOfEndpoints = 200

var _ sdktrace.SpanProcessor = (*Observer)(nil)

// Observer is an observer that can emit RPC metrics.
type Observer struct {
	metricsByEndpoint *MetricsByEndpoint
}

// NewObserver creates a new observer that can emit RPC metrics.
func NewObserver(metricsFactory metrics.Factory, normalizer NameNormalizer) *Observer {
	return &Observer{
		metricsByEndpoint: newMetricsByEndpoint(
			metricsFactory,
			normalizer,
			defaultMaxNumberOfEndpoints,
		),
	}
}

func (o *Observer) OnStart(ctx context.Context, sp sdktrace.ReadWriteSpan) {
	if sp.SpanKind() != trace.SpanKindServer {
		return
	}
	operationName := sp.Name()
	mets := o.metricsByEndpoint.get(operationName)
	mets.RequestProcessing.Update(1)

}

func (o *Observer) OnEnd(sp sdktrace.ReadOnlySpan) {
	operationName := sp.Name()
	if operationName == "" {
		return
	}
	if sp.SpanKind() != trace.SpanKindServer {
		return
	}

	mets := o.metricsByEndpoint.get(operationName)
	latency := sp.EndTime().Sub(sp.StartTime())

	if status := sp.Status(); status.Code == codes.Error {
		mets.RequestCountFailures.Inc(1)
		mets.RequestLatencyFailures.Record(latency)
	} else {
		mets.RequestCountSuccess.Inc(1)
		mets.RequestLatencySuccess.Record(latency)
	}
	for _, attr := range sp.Attributes() {
		if string(attr.Key) == string(semconv.HTTPResponseStatusCodeKey) {
			if attr.Value.Type() == attribute.INT64 {
				mets.recordHTTPStatusCode(attr.Value.AsInt64())
			} else if attr.Value.Type() == attribute.STRING {
				s := attr.Value.AsString()
				if n, err := strconv.Atoi(s); err == nil {
					mets.recordHTTPStatusCode(int64(n))
				}
			}
		}
	}
}

func (*Observer) Shutdown(context.Context) error {
	return nil
}

func (*Observer) ForceFlush(context.Context) error {
	return nil
}
