package registry

import (
	"github.com/zhiyunliu/velocity/metadata"
)

type Options struct {
	cfg      *Config
	Metadata metadata.Metadata
}

type Option func(*Options)
