package logging

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/log"

	"github.com/zhiyunliu/velocity/errors"
	"github.com/zhiyunliu/velocity/middleware"
	"github.com/zhiyunliu/velocity/transport"
)

// Server is an server logging middleware.
func Server(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			var (
				code     int = http.StatusOK
				kind     string
				fullPath string = ctx.Request().Path().FullPath()
			)
			startTime := time.Now()

			if info, ok := transport.FromServerContext(ctx.Context()); ok {
				kind = info.Kind().String()
			}
			ctx.Log().Infof("%s.req %s %s from:%s", kind, ctx.Request().GetMethod(), fullPath, ctx.Request().GetClientIP())

			reply = handler(ctx)
			var err error
			if rerr, ok := reply.(error); ok {
				err = rerr
			}

			if se := errors.FromError(err); se != nil {
				code = se.Code
			}

			level, stack := extractError(err)
			ctx.Log().Logf(level, "%s.resp %s %s %d %s\n%s", kind, ctx.Request().GetMethod(), fullPath, code, time.Since(startTime).String(), stack)
			return
		}
	}
}

// Client is an client logging middleware.
func Client(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			var (
				code      int
				reason    string
				kind      string
				operation string = ctx.Request().Path().FullPath()
			)
			startTime := time.Now()
			if info, ok := transport.FromClientContext(ctx.Context()); ok {
				kind = info.Kind().String()
			}
			reply = handler(ctx)
			var err error
			if rerr, ok := reply.(error); ok {
				err = rerr
			}
			if se := errors.FromError(err); se != nil {
				code = se.Code
			}
			level, stack := extractError(err)
			ctx.Log().Log(level,
				"kind", "client",
				"component", kind,
				"operation", operation,
				"args", extractArgs(ctx.Request().Query()),
				"code", code,
				"reason", reason,
				"stack", stack,
				"latency", time.Since(startTime).Seconds(),
			)
			return
		}
	}
}

// extractArgs returns the string of the req
func extractArgs(req interface{}) string {
	if stringer, ok := req.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", req)
}

// extractError returns the string of the error
func extractError(err error) (log.Level, string) {
	if err != nil {
		return log.LevelError, fmt.Sprintf("%+v", err)
	}
	return log.LevelInfo, ""
}
