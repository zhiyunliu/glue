package xdb

import "sync"

type StreamChan[T any] struct {
	dataChan chan *T
	onceLock sync.Once
}

func NewStreamChan[T any](dataChan chan *T) *StreamChan[T] {
	return &StreamChan[T]{
		dataChan: dataChan,
	}
}

func (r *StreamChan[T]) GetNewRow() any {
	return new(T)
}
func (r *StreamChan[T]) PresentRow(data any) error {
	r.dataChan <- data.(*T)
	return nil
}
func (r *StreamChan[T]) Close() error {
	r.onceLock.Do(func() {
		close(r.dataChan)
	})
	return nil
}
