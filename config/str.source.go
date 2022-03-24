package config

import (
	"github.com/zhiyunliu/golibs/bytesconv"
)

var _ Source = (*strSource)(nil)

type strSource struct {
	content string
}

// NewSource new a file source.
func NewStrSource(content string) Source {
	return &strSource{content: content}
}

func (f *strSource) Load() (kvs []*KeyValue, err error) {

	return []*KeyValue{
		{
			Key:    "str",
			Format: "json",
			Value:  bytesconv.StringToBytes(f.content),
		}}, nil
}

func (f *strSource) Watch() (Watcher, error) {
	return newStrWatcher(f)
}
