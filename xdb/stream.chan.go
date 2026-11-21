package xdb

import "sync"

var _ DataRowReader = (*StreamChan[any])(nil)

type StreamChan[T any] struct {
	dataChan chan *T
	onceLock sync.Once
}

// NewStreamChan 创建一个数据流
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

func (r *StreamChan[T]) GetNext() (any, bool) {
	data, ok := <-r.dataChan
	return data, ok
}

func (r *StreamChan[T]) Close() error {
	r.onceLock.Do(func() {
		close(r.dataChan)
	})
	return nil
}
