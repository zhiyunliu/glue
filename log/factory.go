package log

import "sync"

var (
	builderCache = sync.Map{}
)

func Register(builder Builder) {
	builderCache.Store(builder.Name(), builder)
}

func GetBuilder(builderName string) (Builder, bool) {
	val, ok := builderCache.Load(builderName)
	if !ok {
		return nil, false
	}
	return val.(Builder), true
}
