package context

type Response interface {
	Headers() Header
	Status(int)
	Header(key, val string)
	Write(obj interface{}) error
	WriteBytes([]byte) error
}
