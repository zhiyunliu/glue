package consul

type options struct {
	Addr                string `json:"addr"`
	Heartbeat           bool   `json:"enable_heart_beat"`
	HealthCheck         bool   `json:"enable_health_check"`
	HealthCheckInterval int    `json:"health_check_interval"`
}
