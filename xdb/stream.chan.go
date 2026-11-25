package xdb

import "sync"

var _ DataRowReader = (*StreamChan[any])(nil)

type StreamChan[T any] struct {
	dataChan  chan *T
	closeOnce sync.Once
	stream    chan any
}

// NewStreamChan 创建一个数据流
func NewStreamChan[T any](dataChan chan *T) *StreamChan[T] {
	return &StreamChan[T]{
		dataChan: dataChan,
	}
}

// NewStreamable 创建一个数据流
func NewStreamable[T any]() *StreamChan[T] {
	dataChan := make(chan *T)
	return NewStreamChan(dataChan)
}

func (r *StreamChan[T]) GetNewRow() any {
	return new(T)
}

func (r *StreamChan[T]) PresentRow(data any) error {
	r.dataChan <- data.(*T)
	return nil
}

// GetNext 从数据流中读取下一行，数据流结束时返回 false
func (r *StreamChan[T]) GetNext() (any, bool) {
	data, ok := <-r.dataChan
	return data, ok
}

// GetStreamChan 获取数据流

func (r *StreamChan[T]) GetStreamChan() <-chan *T {
	return r.dataChan
}

func (r *StreamChan[T]) Close() error {
	r.closeOnce.Do(func() {
		close(r.dataChan)
	})
	return nil
}
