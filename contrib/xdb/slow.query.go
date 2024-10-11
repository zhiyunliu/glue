package xdb

import (
	"context"
	"fmt"
	"time"

	"github.com/zhiyunliu/glue/contrib/xdb/implement"
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
	setting.logger.Log(ctx, timeRange.Milliseconds(), fmt.Sprintf("[%s] %s", setting.ConnName, query), implement.Unwrap(args...)...)
}
