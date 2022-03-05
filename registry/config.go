package registry

import "github.com/zhiyunliu/velocity/libs/types"

type Config struct {
	Proto string     `json:"proto"`
	Data  types.XMap `json:"-"`
}
