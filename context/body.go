package context

import "io"

type Body interface {
	Len() int
	io.Reader
	Scan(obj interface{}) error
	Bytes() []byte
}
