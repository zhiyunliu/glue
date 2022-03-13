package context

import "io"

type Body interface {
	io.Reader
	Scan(obj interface{}) error
}
