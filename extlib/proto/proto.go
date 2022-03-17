package proto

import (
	"fmt"
	"strings"
)

//ParseProto 解析协议信息(proto://configname)
func Parse(addr string) (proto string, configname string, err error) {
	addr = strings.TrimSpace(addr)

	parties := strings.SplitN(addr, "://", 2)
	if len(parties) != 2 {
		err = fmt.Errorf("[%s]协议格式错误,正确格式(proto://configname)", addr)
		return
	}
	if proto = parties[0]; proto == "" {
		err = fmt.Errorf("[%s]缺少协议proto,正确格式(proto://configname)", addr)
		return
	}
	if configname = parties[1]; configname == "" {
		err = fmt.Errorf("[%s]缺少configName,正确格式(proto://configname)", addr)
		return
	}
	return
}
