package logger

import (
	"os"
)

type Wrapper struct {
	Logger
	fields map[string]interface{}
}

func NewWrapper(log Logger) *Wrapper {
	return &Wrapper{Logger: log}
}

func (h *Wrapper) Info(args ...interface{}) {
	if !h.Logger.Options().Level.Enabled(InfoLevel) {
		return
	}
	h.Logger.Fields(h.fields).Log(InfoLevel, args...)
}

func (h *Wrapper) Infof(template string, args ...interface{}) {
	if !h.Logger.Options().Level.Enabled(InfoLevel) {
		return
	}
	h.Logger.Fields(h.fields).Logf(InfoLevel, template, args...)
}

func (h *Wrapper) Trace(args ...interface{}) {
	if !h.Logger.Options().Level.Enabled(TraceLevel) {
		return
	}
	h.Logger.Fields(h.fields).Log(TraceLevel, args...)
}

func (h *Wrapper) Tracef(template string, args ...interface{}) {
	if !h.Logger.Options().Level.Enabled(TraceLevel) {
		return
	}
	h.Logger.Fields(h.fields).Logf(TraceLevel, template, args...)
}

func (h *Wrapper) Debug(args ...interface{}) {
	if !h.Logger.Options().Level.Enabled(DebugLevel) {
		return
	}
	h.Logger.Fields(h.fields).Log(DebugLevel, args...)
}

func (h *Wrapper) Debugf(template string, args ...interface{}) {
	if !h.Logger.Options().Level.Enabled(DebugLevel) {
		return
	}
	h.Logger.Fields(h.fields).Logf(DebugLevel, template, args...)
}

func (h *Wrapper) Warn(args ...interface{}) {
	if !h.Logger.Options().Level.Enabled(WarnLevel) {
		return
	}
	h.Logger.Fields(h.fields).Log(WarnLevel, args...)
}

func (h *Wrapper) Warnf(template string, args ...interface{}) {
	if !h.Logger.Options().Level.Enabled(WarnLevel) {
		return
	}
	h.Logger.Fields(h.fields).Logf(WarnLevel, template, args...)
}

func (h *Wrapper) Error(args ...interface{}) {
	if !h.Logger.Options().Level.Enabled(ErrorLevel) {
		return
	}
	h.Logger.Fields(h.fields).Log(ErrorLevel, args...)
}

func (h *Wrapper) Errorf(template string, args ...interface{}) {
	if !h.Logger.Options().Level.Enabled(ErrorLevel) {
		return
	}
	h.Logger.Fields(h.fields).Logf(ErrorLevel, template, args...)
}

func (h *Wrapper) Fatal(args ...interface{}) {
	if !h.Logger.Options().Level.Enabled(FatalLevel) {
		return
	}
	h.Logger.Fields(h.fields).Log(FatalLevel, args...)
	os.Exit(1)
}

func (h *Wrapper) Fatalf(template string, args ...interface{}) {
	if !h.Logger.Options().Level.Enabled(FatalLevel) {
		return
	}
	h.Logger.Fields(h.fields).Logf(FatalLevel, template, args...)
	os.Exit(1)
}

func (h *Wrapper) WithError(err error) *Wrapper {
	fields := copyFields(h.fields)
	fields["error"] = err
	return &Wrapper{Logger: h.Logger, fields: fields}
}

func (h *Wrapper) WithFields(fields map[string]interface{}) *Wrapper {
	nfields := copyFields(fields)
	for k, v := range h.fields {
		nfields[k] = v
	}
	return &Wrapper{Logger: h.Logger, fields: nfields}
}
