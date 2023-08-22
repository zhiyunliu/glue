package context

type Response interface {
	Status(int)
	GetStatusCode() int
	Header(key, val string)
	Write(obj interface{}) error
	WriteBytes([]byte) error
	ContentType() string
	ResponseBytes() []byte
	Redirect(statusCode int, location string)
}
