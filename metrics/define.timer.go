package metrics

import (
	"time"
)

type Timer interface {
	// Records the time passed in.
	Record(time.Duration, ...string)
}

var NoopTimer Timer = noopTimer{}

type noopTimer struct{}

func (noopTimer) Record(time.Duration, ...string) {}
