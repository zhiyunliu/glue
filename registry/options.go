package registry

import "github.com/zhiyunliu/velocity/libs/types"

type Options struct {
	cfg      *Config
	Metadata types.XMap
}

type Option func(*Options)
