package context

import "io"

type Body interface {
	Len() int
	io.Reader
	//Deprecated: Use ScanTo() instead.
	Scan(obj interface{}) error
	ScanTo(obj interface{}) error
	Bytes() []byte
}
