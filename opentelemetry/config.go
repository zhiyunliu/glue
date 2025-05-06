package opentelemetry

type Config struct {
	Insecure     bool   `json:"insecure"`
	Endpoint     string `json:"endpoint"`
	SamplerRate  int64  `json:"sampler_rate"` //0-100
	MetricsProto string `json:"metrics"`
}
