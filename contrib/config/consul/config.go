package consul

type clientConfig struct {
	Addr       string `json:"addr" yaml:"addr"`
	Datacenter string `json:"datacenter" yaml:"datacenter"`
	Namespace  string `json:"namespace" yaml:"namespace"`
	Partition  string `json:"partition" yaml:"partition"`
	Scheme     string `json:"scheme" yaml:"scheme"`
	Token      string `json:"token" yaml:"token"`
}
