package metrics

// Counter is metrics counter.
type Counter interface {
	Inc(...string)
	Add(float64, ...string)
}

var NoopCounter Counter = noopCounter{}

type noopCounter struct{}

func (noopCounter) Inc(...string) {}

func (noopCounter) Add(float64, ...string) {}
