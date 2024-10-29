package xdb

//表达式解析选项
type PropOptions struct {
	UseCache bool
}

type PropOption func(*PropOptions)

// 使用解析缓存
func WithUseCache(use bool) PropOption {
	return func(o *PropOptions) {
		o.UseCache = use
	}
}

//熟悉匹配器
type PropMatcher interface {
	Priority() int
	Pattern() string
	MatchString(string) (PropValuer, bool)
}

type PropValuer interface {
	GetPropName() string
	GetFullfield() string
	GetOper() string
	Build(symbol string, input DBParam, argName string) (string, MissError)
}
