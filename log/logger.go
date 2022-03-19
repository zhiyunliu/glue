package log

var (
	// DefaultLogger logger
	DefaultLogger Logger
)

type Logger interface {
	Name() string

	Log(level Level, content ...interface{})

	Logf(level Level, format string, content ...interface{})

	Infof(format string, content ...interface{})
	Info(content ...interface{})

	Errorf(format string, content ...interface{})
	Error(content ...interface{})

	Debugf(format string, content ...interface{})
	Debug(content ...interface{})

	Fatalf(format string, content ...interface{})
	Fatal(content ...interface{})

	Warnf(format string, v ...interface{})
	Warn(v ...interface{})
}

func String() string {
	return DefaultLogger.Name()
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

func New(name string) Logger {
	return nil
}
