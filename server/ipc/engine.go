package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	gctx "github.com/zhiyunliu/glue/context"
	_ "github.com/zhiyunliu/glue/contrib/engine/gin"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
)

// resolveEngineRoute 通过内存请求复用真实引擎，不创建 HTTP 监听器。
func (s *Server) resolveEngineRoute(ctx context.Context) (func(Request) *Response, error) {
	encode := func(ctx gctx.Context, value interface{}) error {
		response, ok := value.(*Response)
		if !ok {
			response = OK(value)
		}
		data, err := json.Marshal(response)
		if err != nil {
			data, _ = json.Marshal(Fail(CodeEncodeError, "encode_error", err.Error()))
		}
		for k, v := range s.opts.srvCfg.Header {
			ctx.Response().Header(k, v)
		}
		return ctx.Response().WriteBytes(data)
	}
	adapter, err := engine.NewEngine(s.opts.srvCfg.Config.Engine, s.opts.config,
		engine.WithSrvType(s.Type()), func(o *engine.Options) { o.SrvName = s.Name(); o.SvcName = s.ServiceName() },
		engine.WithConfig(s.opts.config), engine.WithLogOptions(s.opts.logOpts), engine.WithRequestDecoder(s.opts.decReq),
		engine.WithResponseEncoder(encode), engine.WithErrorEncoder(func(ctx gctx.Context, err error) { _ = encode(ctx, WrapError(CodeHandlerError, "handler_error", err)) }))
	if err != nil {
		return nil, err
	}
	handler, ok := adapter.GetImpl().(http.Handler)
	if !ok {
		return nil, fmt.Errorf("ipc: engine %q is not http.Handler", s.opts.srvCfg.Config.Engine)
	}
	mids, err := middleware.BuildMiddlewareList(s.opts.srvCfg.Middlewares)
	if err != nil {
		return nil, err
	}
	// 使用额外根节点，避免重启时重复追加配置中间件。
	root := engine.NewRouterGroup("")
	root.Use(mids...)
	root.Use(s.opts.middlewares...)
	root.Children[""] = s.opts.router
	engine.RegistryEngineRoute(adapter, root)
	return func(request Request) *Response {
		path := request.Command()
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://ipc/", bytes.NewReader(request.RawData()))
		if err != nil {
			return WrapError(CodeBadRequest, "bad_request", err)
		}
		req.URL = &url.URL{Scheme: "http", Host: "ipc", Path: path}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		if recorder.Code == http.StatusNotFound || recorder.Code == http.StatusMethodNotAllowed {
			return Fail(CodeUnknownCommand, "unknown_command", "unknown command: "+request.Command())
		}
		var response Response
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			return WrapError(CodeEncodeError, "encode_error", err)
		}
		return &response
	}, nil
}
