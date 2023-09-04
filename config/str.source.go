package config

import (
	"github.com/zhiyunliu/golibs/bytesconv"
)

var _ Source = (*strSource)(nil)

const _strSourceName string = "str"

type strSource struct {
	content string
}

// NewSource new a file source.
func NewStrSource(content string) Source {
	return &strSource{content: content}
}

func (f *strSource) Name() string {
	return _strSourceName
}

func (f *strSource) Path() string {
	return _strSourceName
}

func (f *strSource) Load() (kvs []*KeyValue, err error) {
	return []*KeyValue{
		{
			Key:    _strSourceName,
			Format: "json",
			Value:  bytesconv.StringToBytes(f.content),
		}}, nil
}

func (f *strSource) Watch() (Watcher, error) {
	return newStrWatcher(f)
}
