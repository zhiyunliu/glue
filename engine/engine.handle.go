package engine

import (
	"bytes"
	"time"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/router"
	"github.com/zhiyunliu/golibs/bytesconv"

	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/middleware/otels"
	"github.com/zhiyunliu/glue/middleware/recovery"
)

func RegistryEngineRoute(engine AdapterEngine, router *RouterGroup) {
	defaultMiddlewares := []middleware.Middleware{
		otels.Server(),
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
			kind      = ctx.ServerType()
			fullPath  = ctx.Request().Path().GetURL().Path
			logMethod = ctx.Request().GetMethod()
		)
		logOpts := getLogOptions(ctx)
		startTime := time.Now()
		header := ctx.Request().Header()

		atomicOpts := group.opts.Load()
		if !atomicOpts.GetSkipLog() {
			ctx.Log().Infof("%s.req %s %s from:%s %s", kind, logMethod, fullPath, ctx.Request().GetClientIP(), extractReq(ctx.Request(), logOpts, atomicOpts))
		}
		printSource(ctx.Log(), logOpts, atomicOpts, header)
		printHeader(ctx.Log(), logOpts, atomicOpts, header)

		resp := middleware.Chain(middlewares...)(engineHandler(group, v))(ctx)
		engine.Write(ctx, resp)

		if atomicOpts.GetSkipLog() {
			return
		}
		var err error
		if rerr, ok := resp.(error); ok {
			err = rerr
		}
		code := ctx.Response().GetStatusCode()

		level, errInfo := extractError(err)
		if level == log.LevelError {
			ctx.Log().Logf(level, "%s.resp %s %s %d %s %s %s", kind, logMethod, fullPath, code, time.Since(startTime).String(), extractResp(ctx, logOpts, atomicOpts), errInfo)
		} else {
			ctx.Log().Logf(level, "%s.resp %s %s %d %s %s", kind, logMethod, fullPath, code, time.Since(startTime).String(), extractResp(ctx, logOpts, atomicOpts))
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
func extractReq(req context.Request, logopts *log.Options, ropts *RouterOptions) string {
	res := ""
	if len(req.Query().Values()) > 0 {
		res = req.Query().String()
	}
	if ropts.GetMandatoryLogReq() ||
		(logopts.WithRequest && !(ropts.GetExcludeLogReq() || logopts.IsExclude(req.Path().FullPath()))) {
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

	if ropts.GetMandatoryLogResp() ||
		(logopts.WithResponse && !(ropts.GetExcludeLogResp() || logopts.IsExclude(ctx.Request().Path().FullPath()))) {
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

var (
	SrcHeaders = []string{constants.HeaderSourceIp, constants.HeaderSourceName}
)

func printSource(logger innerLogger, logOpts *log.Options, opts *RouterOptions, header context.Header) {
	if header.IsEmpty() {
		return
	}
	var printSource = false
	//打印服务源
	if logOpts.WithSource != nil {
		printSource = *logOpts.WithSource
	}

	if opts.WithSource != nil {
		printSource = *opts.WithSource
	}

	//不打印，退出
	if !printSource {
		return
	}

	builder := bytes.Buffer{}
	for _, key := range SrcHeaders {
		v := header.Get(key)
		if len(v) <= 0 {
			continue
		}
		builder.WriteString("<")
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(v)
		builder.WriteString(">")
	}
	if builder.Len() > 0 {
		logger.Infof("<reqsrcinfo>:%s", builder.String())
	}
}

func printHeader(logger innerLogger, logOpts *log.Options, opts *RouterOptions, header context.Header) {
	if header.IsEmpty() {
		return
	}

	if opts.GetExcludeHeader() {
		return
	}
	//打印请求头
	var headers = append(make([]constants.HeaderGetter, 0, len(logOpts.WithHeaders)), logOpts.WithHeaders...)
	if len(opts.WithHeaders) > 0 {
		headers = append(headers, opts.WithHeaders...)
	}

	if len(headers) > 0 {
		builder := bytes.Buffer{}
		for _, getter := range headers {
			v := getter.Get(header)
			if len(v) <= 0 {
				continue
			}
			builder.WriteString("<")
			builder.WriteString(getter.Key())
			builder.WriteString("=")
			builder.WriteString(v)
			builder.WriteString(">")
		}
		if builder.Len() > 0 {
			logger.Infof("<reqheader>:%s", builder.String())
		}
	}
}

type LogContext interface {
	LogOptions() *log.Options
}

type innerLogger interface {
	Infof(string, ...any)
}
