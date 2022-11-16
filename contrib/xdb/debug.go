package xdb

import (
	"context"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
	"github.com/zhiyunliu/glue/log"
)

func debugPrint(ctx context.Context, cfg *Config, query string, args ...interface{}) {
	if cfg.Debug {
		logger, ok := log.FromContext(ctx)
		if !ok {
			logger = log.DefaultLogger
		}
		logger.Debugf("sql:%s,args:%v", query, internal.Unwrap(args...))
	}
}
