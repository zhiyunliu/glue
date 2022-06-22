package log

import (
	"sync"

	"github.com/zhiyunliu/golibs/session"
	"github.com/zhiyunliu/golibs/xlog"
)

var (
	DefaultLogger  Logger
	defaultBuilder Builder
	logpool        sync.Pool
)

func init() {
	defaultBuilder = &defaultBuilderWrap{}
	SetBuilder(defaultBuilder)
	logpool = sync.Pool{
		New: func() interface{} {
			return &wraper{}
		},
	}
}

//设置日志的builder
func SetBuilder(builder Builder) {
	if builder == nil {
		return
	}
	defaultBuilder = builder
	DefaultLogger = defaultBuilder.Build(xlog.WithName("default"), xlog.WithSid(session.Create()))
}

type wraper struct {
	xloger xlog.Logger
}

func (l *wraper) Name() string {
	return l.xloger.Name()
}

func (l *wraper) Close() {
	l.xloger.Close()
	logpool.Put(l)
}

func (l *wraper) Log(level Level, args ...interface{}) {
	l.xloger.Log(level, args...)
}

func (l *wraper) Logf(level Level, format string, args ...interface{}) {
	l.xloger.Logf(level, format, args...)
}

func (l *wraper) Info(args ...interface{}) {
	l.Log(LevelInfo, args...)

}

func (l *wraper) Infof(format string, args ...interface{}) {
	l.Logf(LevelInfo, format, args...)

}

func (l *wraper) Error(args ...interface{}) {
	l.Log(LevelError, args...)

}
func (l *wraper) Errorf(format string, args ...interface{}) {
	l.Logf(LevelError, format, args...)

}

func (l *wraper) Debug(args ...interface{}) {
	l.Log(LevelDebug, args...)
}

func (l *wraper) Debugf(format string, args ...interface{}) {
	l.Logf(LevelDebug, format, args...)
}

func (l *wraper) Panic(args ...interface{}) {
	l.Log(LevelPanic, args...)
}

func (l *wraper) Panicf(format string, args ...interface{}) {
	l.Logf(LevelPanic, format, args...)
}

func (l *wraper) Fatalf(format string, args ...interface{}) {
	l.Logf(LevelFatal, format, args...)
}
func (l *wraper) Fatal(args ...interface{}) {
	l.Log(LevelFatal, args...)
}

func (l *wraper) Warnf(format string, args ...interface{}) {
	l.Logf(LevelWarn, format, args...)
}
func (l *wraper) Warn(args ...interface{}) {
	l.Log(LevelWarn, args...)
}

func Debug(args ...interface{}) {
	DefaultLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	DefaultLogger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	DefaultLogger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	DefaultLogger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	DefaultLogger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	DefaultLogger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	DefaultLogger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	DefaultLogger.Errorf(template, args...)
}

func Panic(args ...interface{}) {
	DefaultLogger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	DefaultLogger.Panicf(template, args...)
}

func New(opts ...Option) Logger {
	return defaultBuilder.Build(opts...)
}

func Close() {
	defaultBuilder.Close()
}

func Concurrency(cnt int) {
	xlog.Concurrency(cnt)
}
