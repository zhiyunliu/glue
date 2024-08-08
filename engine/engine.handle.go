package engine

import (
	"bytes"
	"net/http"
	"time"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/router"
	"github.com/zhiyunliu/golibs/bytesconv"

	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/middleware/recovery"
)

func RegistryEngineRoute(engine AdapterEngine, router *RouterGroup) {
	defaultMiddlewares := []middleware.Middleware{
		recovery.Recovery(),
	}
	engine.NoMethod()
	engine.NoRoute()
	execRegistry(engine, router, defaultMiddlewares)
}

func execRegistry(engine AdapterEngine, group *RouterGroup, defaultMiddlewares []middleware.Middleware) {

	groups := group.ServiceGroups
	gmlen := len(group.middlewares)
	mls := make([]middleware.Middleware, len(defaultMiddlewares)+gmlen)
	copy(mls, defaultMiddlewares)

	if gmlen > 0 {
		copy(mls[len(defaultMiddlewares):], group.middlewares)
	}

	for _, v := range groups {
		procHandler(engine, v, mls...)
	}

	for i := range group.Children {
		execRegistry(engine, group.Children[i], mls)
	}
}

func procHandler(engine AdapterEngine, group *RouterWrapper, middlewares ...middleware.Middleware) {
	for method, v := range group.Services {
		engine.Handle(method, group.GetReallyPath(), buildHandler(engine, group, middlewares, v))
	}
	for i := range group.Children {
		procHandler(engine, &RouterWrapper{Group: group.Children[i], opts: group.opts}, middlewares...)
	}
}

func buildHandler(engine AdapterEngine, group *RouterWrapper, middlewares []middleware.Middleware, v *router.Unit) HandlerFunc {
	return func(ctx context.Context) {
		var (
			code      int    = http.StatusOK
			kind             = ctx.ServerType()
			fullPath  string = ctx.Request().Path().GetURL().Path
			logMethod string = ctx.Request().GetMethod()
		)
		logOpts := getLogOptions(ctx)
		startTime := time.Now()
		header := ctx.Request().Header()

		ctx.Log().Infof("%s.req %s %s from:%s %s", kind, logMethod, fullPath, ctx.Request().GetClientIP(), extractReq(ctx.Request(), logOpts, group.opts))

		printSource(ctx.Log(), logOpts, group, header)
		printHeader(ctx.Log(), logOpts, group, header)

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
			ctx.Log().Logf(level, "%s.resp %s %s %d %s %s %s", kind, logMethod, fullPath, code, time.Since(startTime).String(), extractResp(ctx, logOpts, group.opts), errInfo)
		} else {
			ctx.Log().Logf(level, "%s.resp %s %s %d %s %s", kind, logMethod, fullPath, code, time.Since(startTime).String(), extractResp(ctx, logOpts, group.opts))
		}
	}
}

func engineHandler(group *RouterWrapper, unit *router.Unit) middleware.Handler {

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
func extractReq(req context.Request, logopts *log.Options, rotps *RouterOptions) string {
	res := ""
	if len(req.Query().Values()) > 0 {
		res = req.Query().String()
	}
	if logopts.WithRequest && !(rotps.ExcludeLogReq || logopts.IsExclude(req.Path().FullPath())) {
		res += "|"
		res += extractBody(req)
	}
	return res
}

// extractArgs returns the string of the req
func extractBody(req context.Request) string {
	if req.Body().Len() > 0 {
		return bytesconv.BytesToString(req.Body().Bytes())
	}
	return ""
}

func extractResp(ctx context.Context, logopts *log.Options, ropts *RouterOptions) string {
	if logopts.WithResponse && !(ropts.ExcludeLogResp || logopts.IsExclude(ctx.Request().Path().FullPath())) {
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

func getLogOptions(ctx context.Context) *log.Options {
	logCtx, ok := ctx.(LogContext)
	if !ok {
		//todo:应该不会进入该逻辑
		return &log.Options{}
	}
	return logCtx.LogOptions()
}

func printSource(logger innerLogger, logOpts *log.Options, group *RouterWrapper, header context.Header) {
	if header.IsEmpty() {
		return
	}
	var printSource bool = false
	//打印服务源
	if logOpts.WithSource != nil {
		printSource = *logOpts.WithSource
	}
	if group.opts.WithSource != nil {
		printSource = *group.opts.WithSource
	}

	if printSource {
		srcIp := header.Get(constants.HeaderSourceIp)
		srcApp := header.Get(constants.HeaderSourceName)
		if len(srcIp) > 0 || len(srcApp) > 0 {
			logger.Infof("srcinfo:%s:%s,%s:%s", constants.HeaderSourceIp, srcIp, constants.HeaderSourceName, srcApp)
		}
	}
}

func printHeader(logger innerLogger, logOpts *log.Options, group *RouterWrapper, header context.Header) {
	if header.IsEmpty() {
		return
	}
	//打印请求头
	var headers = logOpts.WithHeaders
	if len(group.opts.WithHeaders) > 0 {
		headers = group.opts.WithHeaders
	}

	if len(headers) > 0 {
		builder := bytes.Buffer{}
		for _, key := range headers {
			builder.WriteString(key)
			builder.WriteString(":")
			builder.WriteString(header.Get(key))
			builder.WriteString("\n")
		}
		logger.Infof("header:%s", builder.String())
	}
}

type LogContext interface {
	LogOptions() *log.Options
}

type innerLogger interface {
	Infof(string, ...any)
}
