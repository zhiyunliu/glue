package context

type Response interface {
	Status(int)
	Header(key, val string)
	Write(obj interface{}) error
	WriteBytes([]byte) error
}
