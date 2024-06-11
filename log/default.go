package log

import (
	"context"

	"github.com/zhiyunliu/golibs/session"
	"github.com/zhiyunliu/golibs/xlog"
	_ "github.com/zhiyunliu/golibs/xlog/appenders"
)

var (
	DefaultLogger  Logger
	DefaultBuilder Builder
	//logpool        sync.Pool
)

func init() {
	DefaultBuilder = &defaultBuilderWrap{}
	Register(DefaultBuilder)
	DefaultLogger = DefaultBuilder.Build(context.Background(), xlog.WithName("default"), xlog.WithSid(session.Create()))
}

type Wraper struct {
	Logger xlog.Logger
}

func (l *Wraper) Name() string {
	return l.Logger.Name()
}

func (l *Wraper) Close() {
	l.Logger.Close()
}
func (l *Wraper) Log(level Level, args ...interface{}) {
	l.Logger.Log(level, args...)
}

func (l *Wraper) SessionID() string {
	return l.Logger.SessionID()
}

func (l *Wraper) Logf(level Level, format string, args ...interface{}) {
	l.Logger.Logf(level, format, args...)
}

func (l *Wraper) Info(args ...interface{}) {
	l.Log(LevelInfo, args...)

}

func (l *Wraper) Infof(format string, args ...interface{}) {
	l.Logf(LevelInfo, format, args...)

}

func (l *Wraper) Error(args ...interface{}) {
	l.Log(LevelError, args...)

}
func (l *Wraper) Errorf(format string, args ...interface{}) {
	l.Logf(LevelError, format, args...)

}

func (l *Wraper) Debug(args ...interface{}) {
	l.Log(LevelDebug, args...)
}

func (l *Wraper) Debugf(format string, args ...interface{}) {
	l.Logf(LevelDebug, format, args...)
}

func (l *Wraper) Panic(args ...interface{}) {
	l.Log(LevelPanic, args...)
}

func (l *Wraper) Panicf(format string, args ...interface{}) {
	l.Logf(LevelPanic, format, args...)
}

func (l *Wraper) Fatalf(format string, args ...interface{}) {
	l.Logf(LevelFatal, format, args...)
}
func (l *Wraper) Fatal(args ...interface{}) {
	l.Log(LevelFatal, args...)
}

func (l *Wraper) Warnf(format string, args ...interface{}) {
	l.Logf(LevelWarn, format, args...)
}

func (l *Wraper) Warn(args ...interface{}) {
	l.Log(LevelWarn, args...)
}

func (l *Wraper) Write(p []byte) (n int, err error) {
	l.Log(LevelWarn, string(p))
	return len(p), nil
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

func New(ctx context.Context, opts ...Option) Logger {
	return DefaultBuilder.Build(ctx, opts...)
}

func Close() {
	DefaultBuilder.Close()
}

func Config(opts ...ConfigOption) error {
	return xlog.Config(opts...)
}
