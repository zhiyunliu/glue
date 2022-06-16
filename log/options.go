package log

import "github.com/zhiyunliu/golibs/xlog"

var (
	WithName   = xlog.WithName
	WithSid    = xlog.WithSid
	WithField  = xlog.WithField
	WithFields = xlog.WithFields
)

type Option = xlog.Option
