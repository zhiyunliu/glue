package selector

import (
	"fmt"
)

//selectorResover 定义配置文件转换方法
type selectorResolver interface {
	Name() string
	Resolve() (Selector, error)
}

var _resolvers = make(map[string]selectorResolver)

//Register 注册配置文件适配器
func Register(resolver selectorResolver) {
	name := resolver.Name()
	if _, ok := _resolvers[name]; ok {
		panic(fmt.Errorf("selector: 不能重复注册:%s", name))
	}
	_resolvers[name] = resolver
}

func GetSelector(name string) (Selector, error) {
	resolver, ok := _resolvers[name]
	if !ok {
		return nil, fmt.Errorf("selector: 未知的类型:%s", name)
	}
	return resolver.Resolve()
}
