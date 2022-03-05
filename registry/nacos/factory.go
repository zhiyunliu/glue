package nacos

import "github.com/zhiyunliu/velocity/registry"

type nacosFactory struct {
}

func (f *nacosFactory) Name() string {
	return "nacos"
}

func (f *nacosFactory) Create(opts *registry.Options) (registry.IRegistry, error) {
	return createNacosClient(opts)
}

func init() {
	registry.Register(&nacosFactory{})
}
