package prometheus

type prometheusConfig struct {
	Gateway   *gateway `json:"gateway"`
	Namespace string   `json:"namespace"`
	Job       string   `json:"job"`
}

type gateway struct {
	Addr     string `json:"addr"`
	Interval int    `json:"interval"`
}
