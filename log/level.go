package log

import "github.com/zhiyunliu/golibs/xlog"

type Level = xlog.Level

var (
	LevelAll   xlog.Level = xlog.LevelAll
	LevelDebug xlog.Level = xlog.LevelDebug
	LevelInfo  xlog.Level = xlog.LevelInfo
	LevelWarn  xlog.Level = xlog.LevelWarn
	LevelError xlog.Level = xlog.LevelError
	LevelFatal xlog.Level = xlog.LevelFatal
	LevelOff   xlog.Level = xlog.LevelOff
)
