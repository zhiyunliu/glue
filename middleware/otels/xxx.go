package otels

import (
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

func NewHTTPServer(meter metric.Meter) HTTPServer {
	env := strings.ToLower(os.Getenv(OTelSemConvStabilityOptIn))
	duplicate := env == "http/dup"
	server := HTTPServer{
		duplicate: duplicate,
	}
	server.requestBytesCounter, server.responseBytesCounter, server.serverLatencyMeasure = OldHTTPServer{}.createMeasures(meter)
	if duplicate {
		server.requestBodySizeHistogram, server.responseBodySizeHistogram, server.requestDurationHistogram = CurrentHTTPServer{}.createMeasures(meter)
	}
	return server
}

func xx() {
	if meter == nil {
		return noop.Int64Counter{}, noop.Int64Counter{}, noop.Float64Histogram{}
	}
	var err error
	requestBytesCounter, err := meter.Int64Counter(
		serverRequestSize,
		metric.WithUnit("By"),
		metric.WithDescription("Measures the size of HTTP request messages."),
	)
	handleErr(err)

	responseBytesCounter, err := meter.Int64Counter(
		serverResponseSize,
		metric.WithUnit("By"),
		metric.WithDescription("Measures the size of HTTP response messages."),
	)
	handleErr(err)

	serverLatencyMeasure, err := meter.Float64Histogram(
		serverDuration,
		metric.WithUnit("ms"),
		metric.WithDescription("Measures the duration of inbound HTTP requests."),
	)
	handleErr(err)

	return requestBytesCounter, responseBytesCounter, serverLatencyMeasure
}

type HTTPServer struct {
	duplicate bool

	// Old metrics
	requestBytesCounter  metric.Int64Counter
	responseBytesCounter metric.Int64Counter
	serverLatencyMeasure metric.Float64Histogram

	// New metrics
	requestBodySizeHistogram  metric.Int64Histogram
	responseBodySizeHistogram metric.Int64Histogram
	requestDurationHistogram  metric.Float64Histogram
}

func handleErr(err error) {
	if err != nil {
		otel.Handle(err)
	}
}
