package engine

import (
	"net/http"
	"time"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/router"
	"github.com/zhiyunliu/golibs/bytesconv"

	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/middleware/recovery"
)

func RegistryEngineRoute(engine AdapterEngine, router *RouterGroup, logOpt *log.Options) {
	defaultMiddlewares := []middleware.Middleware{
		//logging.Server(logOpt),
		recovery.Recovery(),
	}
	engine.NoMethod()
	engine.NoRoute()
	execRegistry(engine, router, logOpt, defaultMiddlewares)
}

func execRegistry(engine AdapterEngine, group *RouterGroup, logOpt *log.Options, defaultMiddlewares []middleware.Middleware) {

	groups := group.ServiceGroups
	gmlen := len(group.middlewares)
	mls := make([]middleware.Middleware, len(defaultMiddlewares)+gmlen)
	copy(mls, defaultMiddlewares)

	if gmlen > 0 {
		copy(mls[len(defaultMiddlewares):], group.middlewares)
	}

	for _, v := range groups {
		procHandler(engine, v, logOpt, mls...)
	}

	for i := range group.Children {
		execRegistry(engine, group.Children[i], logOpt, mls)
	}
}

func procHandler(engine AdapterEngine, group *router.Group, opts *log.Options, middlewares ...middleware.Middleware) {
	for method, v := range group.Services {
		engine.Handle(method, group.GetReallyPath(), func(ctx context.Context) {
			var (
				code     int    = http.StatusOK
				kind            = ctx.ServerType()
				fullPath string = ctx.Request().Path().FullPath()
			)
			startTime := time.Now()

			ctx.Log().Infof("%s.req %s %s from:%s %s", kind, ctx.Request().GetMethod(), fullPath, ctx.Request().GetClientIP(), extractReq(opts, ctx.Request()))

			resp := middleware.Chain(middlewares...)(engineHandler(group, v))(ctx)
			engine.Write(ctx, resp)
			var err error
			if rerr, ok := resp.(error); ok {
				err = rerr
			}
			code = ctx.Response().GetStatusCode()
			if se := errors.FromError(err); se != nil {
				code = se.Code
			}

			level, errInfo := extractError(err)
			if level == log.LevelError {
				ctx.Log().Logf(level, "%s.resp %s %s %d %s %s %s", kind, ctx.Request().GetMethod(), fullPath, code, time.Since(startTime).String(), extractResp(opts, ctx), errInfo)
			} else {
				ctx.Log().Logf(level, "%s.resp %s %s %d %s %s", kind, ctx.Request().GetMethod(), fullPath, code, time.Since(startTime).String(), extractResp(opts, ctx))
			}

		})
	}
	for i := range group.Children {
		procHandler(engine, group.Children[i], opts, middlewares...)
	}
}

func engineHandler(group *router.Group, unit *router.Unit) middleware.Handler {

	return func(hctx context.Context) interface{} {

		var resp interface{}
		if unit.Handling != nil {
			resp = unit.Handling.Handle(hctx)
			if _, ok := resp.(*errors.Error); ok {
				return resp
			}
		}
		if group.Handling != nil {
			resp = group.Handling.Handle(hctx)
			if _, ok := resp.(*errors.Error); ok {
				return resp
			}
		}
		handleResp := unit.Handle.Handle(hctx)
		if unit.Handled != nil {
			resp = unit.Handled.Handle(hctx)
			if _, ok := resp.(*errors.Error); ok {
				return resp
			}
		}
		if group.Handled != nil {
			resp = group.Handled.Handle(hctx)
			if _, ok := resp.(*errors.Error); ok {
				return resp
			}
		}
		return handleResp
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

func extractResp(opts *log.Options, ctx context.Context) string {
	if opts.WithResponse && !opts.IsExclude(ctx.Request().Path().FullPath()) {
		return bytesconv.BytesToString(ctx.Response().ResponseBytes())
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
