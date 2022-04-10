package logging

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/gel/log"
	"github.com/zhiyunliu/golibs/bytesconv"

	"github.com/zhiyunliu/gel/errors"
	"github.com/zhiyunliu/gel/middleware"
)

// Server is an server logging middleware.
func Server(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			var (
				code     int    = http.StatusOK
				kind            = ctx.ServerType()
				fullPath string = ctx.Request().Path().FullPath()
			)
			startTime := time.Now()

			ctx.Log().Infof("%s.req %s %s from:%s %s", kind, ctx.Request().GetMethod(), fullPath, ctx.Request().GetClientIP(), extractReq(ctx.Request()))

			reply = handler(ctx)
			var err error
			if rerr, ok := reply.(error); ok {
				err = rerr
			}

			if se := errors.FromError(err); se != nil {
				code = se.Code
			}

			level, stack := extractError(err)
			if level == log.LevelError {
				ctx.Log().Logf(level, "%s.resp %s %s %d %s\n%s", kind, ctx.Request().GetMethod(), fullPath, code, time.Since(startTime).String(), stack)
			} else {
				ctx.Log().Logf(level, "%s.resp %s %s %d %s %s", kind, ctx.Request().GetMethod(), fullPath, code, time.Since(startTime).String(), extractResp(reply))
			}
			return
		}
	}
}

// extractArgs returns the string of the req
func extractReq(req context.Request) string {
	var query, body string
	if len(req.Query().SMap()) > 0 {
		query = req.Query().String()
	}
	if req.Body().Len() > 0 {
		body = bytesconv.BytesToString(req.Body().Bytes())
	}
	return fmt.Sprintf("query:%s,body:%s", query, body)
}

func extractResp(resp interface{}) string {
	if stringer, ok := resp.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", resp)
}

// extractError returns the string of the error
func extractError(err error) (log.Level, string) {
	if err != nil {
		return log.LevelError, fmt.Sprintf("%+v", err)
	}
	return log.LevelInfo, ""
}
