package context

type Response interface {
	Headers() Header
	Status() int
	Header(key, val string)
}
