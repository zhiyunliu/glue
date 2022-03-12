package log

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
	if !h.Logger.Enabled(LevelInfo) {
		return
	}
	h.Logger.Fields(h.fields).Log(LevelInfo, args...)
}

func (h *Wrapper) Infof(template string, args ...interface{}) {
	if !h.Logger.Enabled(LevelInfo) {
		return
	}
	h.Logger.Fields(h.fields).Logf(LevelInfo, template, args...)
}

func (h *Wrapper) Trace(args ...interface{}) {
	if !h.Logger.Enabled(LevelTrace) {
		return
	}
	h.Logger.Fields(h.fields).Log(LevelTrace, args...)
}

func (h *Wrapper) Tracef(template string, args ...interface{}) {
	if !h.Logger.Enabled(LevelTrace) {
		return
	}
	h.Logger.Fields(h.fields).Logf(LevelTrace, template, args...)
}

func (h *Wrapper) Debug(args ...interface{}) {
	if !h.Logger.Enabled(LevelDebug) {
		return
	}
	h.Logger.Fields(h.fields).Log(LevelDebug, args...)
}

func (h *Wrapper) Debugf(template string, args ...interface{}) {
	if !h.Logger.Enabled(LevelDebug) {
		return
	}
	h.Logger.Fields(h.fields).Logf(LevelDebug, template, args...)
}

func (h *Wrapper) Warn(args ...interface{}) {
	if !h.Logger.Enabled(LevelWarn) {
		return
	}
	h.Logger.Fields(h.fields).Log(LevelWarn, args...)
}

func (h *Wrapper) Warnf(template string, args ...interface{}) {
	if !h.Logger.Enabled(LevelWarn) {
		return
	}
	h.Logger.Fields(h.fields).Logf(LevelWarn, template, args...)
}

func (h *Wrapper) Error(args ...interface{}) {
	if !h.Logger.Enabled(LevelError) {
		return
	}
	h.Logger.Fields(h.fields).Log(LevelError, args...)
}

func (h *Wrapper) Errorf(template string, args ...interface{}) {
	if !h.Logger.Enabled(LevelError) {
		return
	}
	h.Logger.Fields(h.fields).Logf(LevelError, template, args...)
}

func (h *Wrapper) Fatal(args ...interface{}) {
	if !h.Logger.Enabled(LevelFatal) {
		return
	}
	h.Logger.Fields(h.fields).Log(LevelFatal, args...)
	os.Exit(1)
}

func (h *Wrapper) Fatalf(template string, args ...interface{}) {
	if !h.Logger.Enabled(LevelFatal) {
		return
	}
	h.Logger.Fields(h.fields).Logf(LevelFatal, template, args...)
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
