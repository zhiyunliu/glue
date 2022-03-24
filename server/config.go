package server

import (
	"github.com/zhiyunliu/golibs/xtypes"
)

type Middleware struct {
	Name string
	Data MiddlewareConfig
}

type MiddlewareConfig struct {
	Proto string
	Data  []byte
}

type Header xtypes.SMap
