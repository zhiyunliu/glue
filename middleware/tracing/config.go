package tracing

//	{"provider":"skywalking","propagator":"propagator"}
type Config struct {
	Provider   string `json:"provider" yaml:"provider"`
	Propagator string `json:"propagator" yaml:"propagator"`
}
