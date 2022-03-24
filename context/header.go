package context

type Header interface {
	Get(name string) string
	Set(name, val string)
	Scan(obj interface{}) error
}
