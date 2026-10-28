package defaultipc

import "github.com/zhiyunliu/glue/log"

type loggerWrap struct {
	logger log.Logger
}

func (l *loggerWrap) Printf(format string, args ...any) {
	l.logger.Infof(format, args...)
}
