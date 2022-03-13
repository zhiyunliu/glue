package context

type Header interface {
	Get(name string) string
	Scan(obj interface{}) error
}
