package xdb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
	"github.com/zhiyunliu/glue/log"
)

func debugPrint(ctx context.Context, setting *Setting, query string, args ...interface{}) {
	if setting.Cfg.Debug {
		logger, ok := log.FromContext(ctx)
		if !ok {
			logger = log.DefaultLogger
		}
		builder := strings.Builder{}
		args = internal.Unwrap(args...)
		idx := 1
		for _, v := range args {
			if na, ok := v.(sql.NamedArg); ok {
				builder.WriteString(fmt.Sprintf("@%s = %v\n", na.Name, na.Value))
				continue
			}
			if na, ok := v.(*sql.NamedArg); ok {
				builder.WriteString(fmt.Sprintf("@%s = %v\n", na.Name, na.Value))
				continue
			}
			builder.WriteString(fmt.Sprintf("@p%d = %v\n", idx, v))
			idx++
		}

		logger.Debugf("sql:%s,args:%s", query, builder.String())
	}
}
