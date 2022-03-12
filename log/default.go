package log

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

func init() {
	global.SetLogger(DefaultLogger)

}

var global = &loggerAppliance{}

// loggerAppliance is the proxy of `Logger` to
// make logger change will affect all sub-logger.
type loggerAppliance struct {
	lock sync.Mutex
	Logger
	wrapper *Wrapper
}

func (a *loggerAppliance) SetLogger(in Logger) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.Logger = in
	a.wrapper = NewWrapper(a.Logger)
}

func (a *loggerAppliance) GetLogger() Logger {
	return a.Logger
}

type defaultLogger struct {
	sync.RWMutex
	opts *Options
}

func (l *defaultLogger) String() string {
	return "default"
}

func (l *defaultLogger) Enabled(level Level) bool {
	return l.opts.Level.Enabled(level)
}

func (l *defaultLogger) Fields(fields map[string]interface{}) Logger {
	l.Lock()
	l.opts.Fields = copyFields(fields)
	l.Unlock()
	return l
}

func copyFields(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// logCallerfilePath returns a package/file:line description of the caller,
// preserving only the leaf directory name and file name.
func logCallerfilePath(loggingFilePath string) string {
	// To make sure we trim the path correctly on Windows too, we
	// counter-intuitively need to use '/' and *not* os.PathSeparator here,
	// because the path given originates from Go stdlib, specifically
	// runtime.Caller() which (as of Mar/17) returns forward slashes even on
	// Windows.
	//
	// See https://github.com/golang/go/issues/3335
	// and https://github.com/golang/go/issues/18151
	//
	// for discussion on the issue on Go side.
	idx := strings.LastIndexByte(loggingFilePath, '/')
	if idx == -1 {
		return loggingFilePath
	}
	idx = strings.LastIndexByte(loggingFilePath[:idx], '/')
	if idx == -1 {
		return loggingFilePath
	}
	return loggingFilePath[idx+1:]
}

func (l *defaultLogger) Log(level Level, v ...interface{}) {
	l.logf(level, "", v...)
}

func (l *defaultLogger) Logf(level Level, format string, v ...interface{}) {
	l.logf(level, format, v...)
}

func (l *defaultLogger) logf(level Level, format string, v ...interface{}) {
	if !l.opts.Level.Enabled(level) {
		return
	}

	l.RLock()
	fields := copyFields(l.opts.Fields)
	l.RUnlock()

	fields["level"] = level.String()

	if _, file, line, ok := runtime.Caller(l.opts.CallerSkipCount); ok {
		fields["file"] = fmt.Sprintf("%s:%d", logCallerfilePath(file), line)
	}

	keys := make([]string, 0, len(fields))

	sort.Strings(keys)
	metadata := ""

	for i, k := range keys {
		if i == 0 {
			metadata += fmt.Sprintf("%s:%v", k, fields[k])
		} else {
			metadata += fmt.Sprintf(" %s:%v", k, fields[k])
		}
	}

	var name string
	if l.opts.Name != "" {
		name = "[" + l.opts.Name + "]"
	}
	t := time.Now().Format("2006-01-02 15:04:05.000")
	//fmt.Printf("%s\n", t)
	//fmt.Printf("%s\n", name)
	//fmt.Printf("%s\n", metadata)
	//fmt.Printf("%v\n", rec.Message)
	logStr := ""
	if name == "" {
		logStr = fmt.Sprintf("%s %s %v\n", t, metadata, buildMsg(format, v...))
	} else {
		logStr = fmt.Sprintf("%s %s %s %v\n", name, t, metadata, buildMsg(format, v...))
	}
	_, err := l.opts.Out.Write([]byte(logStr))
	if err != nil {
		log.Printf("log [Logf] write error: %s \n", err.Error())
	}

}

// NewLogger builds a new logger based on options
func NewLogger(opts ...Option) Logger {
	// Default options
	options := &Options{
		Level:           LevelInfo,
		Fields:          make(map[string]interface{}),
		Out:             os.Stderr,
		CallerSkipCount: 3,
		Context:         context.Background(),
		Name:            "",
	}

	for i := range opts {
		opts[i](options)
	}

	l := &defaultLogger{opts: options}
	return l
}

func buildMsg(format string, v ...interface{}) string {
	if format == "" {
		return fmt.Sprint(v...)
	}
	return fmt.Sprintf(format, v...)
}
