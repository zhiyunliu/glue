package xdb

import "sync"

var (
	loggerCache = sync.Map{}
)

type Logger interface {
	Name() string
	Log(args ...interface{})
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
