package xdb

import (
	"context"
	"time"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
)

// 打印慢sql
func printSlowQuery(ctx context.Context, setting *Setting, timeRange time.Duration, query string, args ...interface{}) {
	if !setting.Cfg.ShowQueryLog {
		return
	}
	if setting.logger == nil {
		return
	}

	if setting.slowThreshold <= 0 || timeRange < setting.slowThreshold {
		return
	}
	setting.logger.Log(ctx, timeRange.Milliseconds(), query, internal.Unwrap(args...))
}
