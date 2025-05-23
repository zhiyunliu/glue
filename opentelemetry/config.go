package opentelemetry

type Config struct {
	//Insecure indicates whether to use insecure connection
	// for the OpenTelemetry collector.
	Insecure bool `json:"insecure"`
	// TraceEndpoint is the endpoint for the OpenTelemetry collector.
	TraceEndpoint string `json:"trace_endpoint"`
	// TraceSampleRate is the sample rate for traces.
	// It should be a value between 0 and 100.
	TraceSampleRate int64 `json:"trace_sample_rate"` //0-100
	// MetricsProvider is the provider for metrics.
	// It can be "prometheus" or "statsd".
	MetricsProvider string `json:"metrics_provider"`
}
