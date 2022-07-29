package log

type Logger interface {
	Name() string

	Log(level Level, args ...interface{})
	Logf(level Level, format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Panic(args ...interface{})
	Panicf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	Warn(v ...interface{})
	Warnf(format string, v ...interface{})

	Write(p []byte) (n int, err error)
	Close()
}
