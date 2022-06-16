package p2c

import "github.com/zhiyunliu/glue/selector"

type p2cResolver struct{}

func (p2cResolver) Name() string {
	return Name
}

func (p2cResolver) Resolve() (selector.Selector, error) {
	return NewBuilder().Build(), nil
}

func init() {
	selector.Register(&p2cResolver{})
}
