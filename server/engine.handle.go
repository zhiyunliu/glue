package server

import (
	"github.com/zhiyunliu/gel/context"

	"github.com/zhiyunliu/gel/errors"
	"github.com/zhiyunliu/gel/middleware"
	"github.com/zhiyunliu/gel/middleware/logging"
	"github.com/zhiyunliu/gel/middleware/recovery"
	"github.com/zhiyunliu/gel/reflect"
)

type AdapterEngine interface {
	NoMethod()
	NoRoute()
	Handle(method string, path string, callfunc HandlerFunc)
	Write(ctx context.Context, resp interface{})
}
type HandlerFunc func(context.Context)

func RegistryEngineRoute(engine AdapterEngine, router *RouterGroup) {
	defaultMiddlewares := []middleware.Middleware{
		logging.Server(nil),
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
		execRegistry(engine, group.Children[i], defaultMiddlewares)
	}
}

func procHandler(engine AdapterEngine, group *reflect.ServiceGroup, middlewares ...middleware.Middleware) {
	for method, v := range group.Services {
		engine.Handle(method, group.GetReallyPath(), func(ctx context.Context) {
			resp := middleware.Chain(middlewares...)(engineHandler(group, v))(ctx)
			engine.Write(ctx, resp)
		})
	}
	for i := range group.Children {
		procHandler(engine, group.Children[i], middlewares...)
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
