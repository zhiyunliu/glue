package xdb

import (
	"context"
	"sync"
)

var (
	loggerCache = sync.Map{}
)

type Logger interface {
	Name() string
	Log(ctx context.Context, elapsed int64, sql string, args ...interface{})
}

func RegistryLogger(logger Logger) {
	loggerCache.Store(logger.Name(), logger)
}

func GetLogger(loggerName string) (Logger, bool) {
	val, ok := loggerCache.Load(loggerName)
	if !ok {
		return nil, false
	}
	return val.(Logger), true
}
