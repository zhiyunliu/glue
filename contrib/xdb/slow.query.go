package xdb

import (
	"time"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
)

//打印慢sql
func printSlowQuery(cfg *Config, timeRange time.Duration, query string, args ...interface{}) {
	if !cfg.ShowQueryLog {
		return
	}
	if cfg.logger == nil {
		return
	}

	if timeRange < cfg.slowThreshold {
		return
	}

	cfg.logger.Log(map[string]interface{}{
		"time":  timeRange.Milliseconds(),
		"query": query,
		"args":  internal.Unwrap(args...),
	})
}
