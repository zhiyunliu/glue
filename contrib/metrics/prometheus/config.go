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

// GetAddr returns the pushgateway address.
func (g *gateway) GetAddr() string {
	return g.Addr
}

// GetInterval returns the pushgateway interval.
func (g *gateway) GetInterval() int {
	if g.Interval == 0 {
		return 15 // default interval 15s
	}
	return g.Interval
}
