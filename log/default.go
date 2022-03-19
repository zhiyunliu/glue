package log

import (
	"sync"
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
}

func (a *loggerAppliance) SetLogger(in Logger) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.Logger = in
}

func (a *loggerAppliance) GetLogger() Logger {
	return a.Logger
}
