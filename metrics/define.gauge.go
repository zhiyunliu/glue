package metrics

// Gauge is metrics gauge.
type Gauge interface {
	Set(value float64, lbvs ...string)
	Add(delta float64, lbvs ...string)
	Sub(delta float64, lbvs ...string)
}

var NoopGauge Gauge = noopGauge{}

type noopGauge struct{}

func (noopGauge) Set(float64, ...string) {}

func (noopGauge) Add(float64, ...string) {}

func (noopGauge) Sub(float64, ...string) {}
