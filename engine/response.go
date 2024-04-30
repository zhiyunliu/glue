package engine

import "github.com/zhiyunliu/golibs/xtypes"

// ResponseWriter ...
type ResponseWriter interface {
	Status() int
	Size() int
	Written() bool
	WriteHeader(code int)
	Header() xtypes.SMap
	Write(p []byte) (n int, err error)
	WriteString(string) (int, error)
	Flush() error
}
