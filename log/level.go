package log

import (
	"os"
)

type Level int8

const (
	LevelTrace Level = 0
	LevelDebug Level = 1
	LevelInfo  Level = 2
	LevelWarn  Level = 3
	LevelError Level = 4
	LevelFatal Level = 5
	LevelAll   Level = 6
)

func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "trace"
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	}
	return "All"
}

func (l Level) Short() string {
	switch l {
	case LevelTrace:
		return "t"
	case LevelDebug:
		return "d"
	case LevelInfo:
		return "i"
	case LevelWarn:
		return "w"
	case LevelError:
		return "e"
	case LevelFatal:
		return "f"
	}
	return "a"
}

// Enabled returns true if the given level is at or above this level.
func (l Level) Enabled(lvl Level) bool {
	return lvl >= l
}

func Trace(args ...interface{}) {
	DefaultLogger.Log(LevelTrace, args...)
}

func Tracef(template string, args ...interface{}) {
	DefaultLogger.Logf(LevelTrace, template, args...)
}

func Debug(args ...interface{}) {
	DefaultLogger.Log(LevelDebug, args...)
}

func Debugf(template string, args ...interface{}) {
	DefaultLogger.Logf(LevelDebug, template, args...)
}

func Info(args ...interface{}) {
	DefaultLogger.Log(LevelInfo, args...)
}

func Infof(template string, args ...interface{}) {
	DefaultLogger.Logf(LevelInfo, template, args...)
}

func Warn(args ...interface{}) {
	DefaultLogger.Log(LevelWarn, args...)
}

func Warnf(template string, args ...interface{}) {
	DefaultLogger.Logf(LevelWarn, template, args...)
}

func Error(args ...interface{}) {
	DefaultLogger.Log(LevelError, args...)
}

func Errorf(template string, args ...interface{}) {
	DefaultLogger.Logf(LevelError, template, args...)
}

func Fatal(args ...interface{}) {
	DefaultLogger.Log(LevelFatal, args...)
	os.Exit(1)
}

func Fatalf(template string, args ...interface{}) {
	DefaultLogger.Logf(LevelFatal, template, args...)
	os.Exit(1)
}

// // Returns true if the given level is at or lower the current logger level
// func V(lvl Level, log Logger) bool {
// 	l := DefaultLogger
// 	if log != nil {
// 		l = log
// 	}
// 	return l.Options().Level <= lvl
// }
