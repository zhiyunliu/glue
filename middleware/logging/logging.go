package logging

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/velocity/context"
	"github.com/zhiyunliu/velocity/log"

	"github.com/zhiyunliu/velocity/errors"
	"github.com/zhiyunliu/velocity/middleware"
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
	data := map[string]interface{}{}
	if len(req.Query().SMap()) > 0 {
		data["query"] = req.Query().SMap()
	}
	if req.Body().Len() > 0 {
		data["body"] = bytesconv.BytesToString(req.Body().Bytes())
	}
	reqBytes, _ := json.Marshal(data)
	return bytesconv.BytesToString(reqBytes)
}

func extractResp(req interface{}) string {
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
