package metrics

// Observer is metrics observer.
type Histogram interface {
	Record(float64, ...string)
}

var NoopHistogram Histogram = noopHistogram{}

type noopHistogram struct{}

func (noopHistogram) Record(float64, ...string) {}
