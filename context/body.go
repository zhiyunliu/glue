package context

import "io"

type Body interface {
	Len() int
	io.Reader
	ScanTo(obj interface{}) error
	Bytes() []byte
}
