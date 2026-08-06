package context

import "time"

type Response interface {
	StatusCode(int)
	GetStatusCode() int
	GetHeader(key string) string
	Header(key, val string)
	Write(obj interface{}) error
	WriteBytes([]byte) error
	ContentType() string
	ResponseBytes() []byte
	Size() int
	Redirect(statusCode int, location string)
	Flush() error
}

type WriteDeadlineSetter interface {
	SetWriteDeadline(time.Time) error
}
