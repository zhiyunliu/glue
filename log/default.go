package log

import (
	"github.com/zhiyunliu/golibs/session"
	"github.com/zhiyunliu/golibs/xlog"
)

func init() {
	DefaultLogger = &wraper{
		xloger: xlog.New(xlog.WithName("default"), xlog.WithSid(session.Create())),
	}
}

var DefaultLogger Logger

type wraper struct {
	xloger xlog.Logger
}

func (l *wraper) Name() string {
	return l.xloger.Name()
}

func (l *wraper) Close() {
	l.xloger.Close()
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

func New(opts ...Option) Logger {
	return &wraper{
		xloger: xlog.GetLogger(opts...),
	}
}

func Close() {
	xlog.Close()
}
