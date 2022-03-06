package context

type Body interface {
	Scan(obj interface{}) error
}
