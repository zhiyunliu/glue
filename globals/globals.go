package globals

import (
	"sync"

	"github.com/zhiyunliu/velocity/logger"
)

var glogger logger.Logger
var onceLock sync.Once

func Logger() logger.Logger {
	onceLock.Do(func() {
		glogger = logger.GetLogger("")

	})
	return glogger
}
