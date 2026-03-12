package context

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
