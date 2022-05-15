package wrr

import "github.com/zhiyunliu/gel/selector"

type wrrResolver struct{}

func (wrrResolver) Name() string {
	return Name
}

func (wrrResolver) Resolve() (selector.Selector, error) {
	return NewBuilder().Build(), nil
}

func init() {
	selector.Register(&wrrResolver{})
}
