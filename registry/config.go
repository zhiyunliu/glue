package registry

import "github.com/zhiyunliu/velocity/extlib/xtypes"

type Config struct {
	Proto string      `json:"proto"`
	Data  xtypes.XMap `json:"-"`
}
