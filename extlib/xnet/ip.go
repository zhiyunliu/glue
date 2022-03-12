package xnet

import (
	"net"
	"strings"
)

const StaticLocalIP = "127.0.0.1"

// GetLocalIP 获取IP地址
func GetLocalIP(masks ...string) string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return StaticLocalIP
	}
	var ipLst []string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			ipLst = append(ipLst, ipnet.IP.String())
		}
	}
	if len(masks) == 0 && len(ipLst) > 0 {
		return ipLst[0]
	}
	for _, ip := range ipLst {
		for _, m := range masks {
			if strings.HasPrefix(ip, m) {
				return ip
			}
		}
	}
	return StaticLocalIP
}
