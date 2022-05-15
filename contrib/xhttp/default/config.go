package grpc

import "github.com/zhiyunliu/gel/config"

type setting struct {
	Name                string        `json:"-"`
	Balancer            string        `json:"balancer"` //selector.Selector
	ConnTimeout         int           `json:"conn_timeout"`
	CertFile            string        `json:"cert_file"`
	KeyFile             string        `json:"file_file"`
	CaFile              string        `json:"ca_file"`
	ProxyURL            string        `json:"proxy_url"`
	KeepaliveTimeout    int           `json:"keep_alive_timeout"`
	MaxIdleConns        int           `json:"max_idle_conns"`
	IdleConnTimeout     int           `json:"idle_conn_timeout"`
	TLSHandshakeTimeout int           `json:"tls_handshake_timeout"`
	Config              config.Config `json:"-"`
}
