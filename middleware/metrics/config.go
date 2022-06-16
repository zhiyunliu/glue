package metrics

//	{"counter":"aaaa","observer":"bbbb"}
type Config struct {
	CounterName  string `json:"counter" yaml:"counter"`
	ObserverName string `json:"observer" yaml:"observer"`
}
