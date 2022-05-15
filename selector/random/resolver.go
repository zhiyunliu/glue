package random

import "github.com/zhiyunliu/gel/selector"

type randomResolver struct{}

func (randomResolver) Name() string {
	return Name
}

func (randomResolver) Resolve() (selector.Selector, error) {
	return NewBuilder().Build(), nil
}

func init() {
	selector.Register(&randomResolver{})
}
