package logging

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/golibs/bytesconv"

	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/middleware"
)

// Server is an server logging middleware.
func Server(opts *log.Options) middleware.Middleware {

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context) (reply interface{}) {
			var (
				code     int    = http.StatusOK
				kind            = ctx.ServerType()
				fullPath string = ctx.Request().Path().FullPath()
			)
			startTime := time.Now()

			ctx.Log().Infof("%s.req %s %s from:%s %s", kind, ctx.Request().GetMethod(), fullPath, ctx.Request().GetClientIP(), extractReq(opts, ctx.Request()))

			reply = handler(ctx)
			var err error
			if rerr, ok := reply.(error); ok {
				err = rerr
			}

			if se := errors.FromError(err); se != nil {
				code = se.Code
			}

			level, errInfo := extractError(err)
			if level == log.LevelError {
				ctx.Log().Logf(level, "%s.resp %s %s %d %s %s %s", kind, ctx.Request().GetMethod(), fullPath, code, time.Since(startTime).String(), extractResp(opts, ctx.Request(), reply), errInfo)
			} else {
				ctx.Log().Logf(level, "%s.resp %s %s %d %s %s", kind, ctx.Request().GetMethod(), fullPath, code, time.Since(startTime).String(), extractResp(opts, ctx.Request(), reply))
			}
			return
		}
	}
}

// extractArgs returns the string of the req
func extractReq(opts *log.Options, req context.Request) string {
	res := ""
	if len(req.Query().Values()) > 0 {
		res = req.Query().String()
	}
	if opts.WithRequest && !opts.IsExclude(req.Path().FullPath()) {
		res += "|"
		res += extractBody(opts, req)
	}
	return res
}

// extractArgs returns the string of the req
func extractBody(opts *log.Options, req context.Request) string {
	if req.Body().Len() > 0 {
		return bytesconv.BytesToString(req.Body().Bytes())
	}
	return ""
}

func extractResp(opts *log.Options, req context.Request, resp interface{}) string {
	if opts.WithResponse && !opts.IsExclude(req.Path().FullPath()) {
		if stringer, ok := resp.(fmt.Stringer); ok {
			return stringer.String()
		}
		return fmt.Sprintf("%+v", resp)
	}
	return ""
}

// extractError returns the string of the error
func extractError(err error) (log.Level, string) {
	if err != nil {
		return log.LevelError, err.Error()
	}
	return log.LevelInfo, ""
}
