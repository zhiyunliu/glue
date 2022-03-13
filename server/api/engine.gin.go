package api

import (
	"github.com/zhiyunliu/velocity/context"

	"github.com/gin-gonic/gin"
	"github.com/zhiyunliu/velocity/errors"
	"github.com/zhiyunliu/velocity/middleware"
	"github.com/zhiyunliu/velocity/middleware/logging"
	"github.com/zhiyunliu/velocity/middleware/recovery"
	"github.com/zhiyunliu/velocity/reflect"
	"github.com/zhiyunliu/velocity/server"
)

func (e *Server) registryEngineRoute() {
	engine := e.opts.handler.(*gin.Engine)
	defaultMiddlewares := []middleware.Middleware{
		logging.Server(nil),
		recovery.Recovery(),
	}

	e.execRegistry(engine, e.opts.router, defaultMiddlewares)
}

func (e *Server) execRegistry(engine *gin.Engine, group *server.RouterGroup, defaultMiddlewares []middleware.Middleware) {

	groups := group.ServiceGroups
	mls := make([]middleware.Middleware, len(defaultMiddlewares)+len(groups))
	copy(mls, defaultMiddlewares)
	copy(mls[len(defaultMiddlewares):], defaultMiddlewares)
	for _, v := range groups {

		e.procHandler(engine, v, mls...)
	}

	for i := range group.Children {
		e.execRegistry(engine, group.Children[i], defaultMiddlewares)
	}
}

func (e *Server) procHandler(engine *gin.Engine, group *reflect.ServiceGroup, middlewares ...middleware.Middleware) {
	for method, v := range group.Services {
		engine.Handle(method, group.GetReallyPath(), func(ctx *gin.Context) {
			var ginCtx *server.GinContext = &server.GinContext{
				Gctx: ctx,
			}
			res := middleware.Chain(middlewares...)(engineHandler(group, v))(ginCtx)
			e.opts.enc(ginCtx, res)
		})
	}
	for i := range group.Children {
		e.procHandler(engine, group.Children[i])
	}
}

func engineHandler(group *reflect.ServiceGroup, unit *reflect.ServiceUnit) middleware.Handler {

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
