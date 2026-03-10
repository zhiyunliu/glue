package log

type Logger interface {
	Name() string

	SessionID() string

	Log(level Level, args ...any)
	Logf(level Level, format string, args ...any)

	Info(args ...any)
	Infof(format string, args ...any)

	Error(args ...any)
	Errorf(format string, args ...any)

	Debug(args ...any)
	Debugf(format string, args ...any)

	Panic(args ...any)
	Panicf(format string, args ...any)

	Fatal(args ...any)
	Fatalf(format string, args ...any)

	Warn(v ...any)
	Warnf(format string, v ...any)

	Chain(level Level, msg string, opts ...EventOption)

	Write(p []byte) (n int, err error)
	Close()
}
