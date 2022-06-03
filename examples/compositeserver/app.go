package main

import (
	"net/http"
	"time"

	"github.com/zhiyunliu/gel"
	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/gel/examples/compositeserver/handles"
	"github.com/zhiyunliu/gel/middleware/tracing"
	"github.com/zhiyunliu/gel/transport"
	"github.com/zhiyunliu/gel/xhttp"

	"github.com/zhiyunliu/gel/server/api"
	"github.com/zhiyunliu/gel/server/cron"
	"github.com/zhiyunliu/gel/server/mqc"
	"github.com/zhiyunliu/gel/server/rpc"
	"github.com/zhiyunliu/golibs/xtypes"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var Name = "compositeserver"

func init() {

	srvOpt := gel.Server(
		apiserver(),
		mqcserver(),
		cronserver(),
		rpcserver(),
	)
	opts = append(opts, srvOpt, gel.LogConcurrency(1))
	setTracerProvider("http://127.0.0.1:14268/api/traces")
}

// Set global trace provider
func setTracerProvider(url string) error {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return err
	}
	tp := tracesdk.NewTracerProvider(
		// Set the sampling rate based on the parent span to 100%
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.AlwaysSample())),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(Name),
			attribute.String("env", "dev"),
		)),
	)
	otel.SetTracerProvider(tp)
	return nil
}

func apiserver() transport.Server {
	apiSrv := api.New("apiserver")
	apiSrv.Use(tracing.Server(tracing.WithPropagator(propagation.TraceContext{}), tracing.WithTracerProvider(otel.GetTracerProvider())))
	apiSrv.Handle("/log", handles.NewLogDemo())
	apiSrv.Handle("/demoapi", func(ctx context.Context) interface{} {
		ctx.Log().Debug("demo")

		body, err := gel.Http().GetHttp().Swap(ctx, "http://apiserver/log/info")
		if err != nil {
			ctx.Log().Error("gel.Http().GetHttp().Swap:", err)
		}
		// body, err := gel.RPC().GetRPC().Swap(ctx, "grpc://compositeserver/demorpc", xrpc.WithWaitForReady(false))
		// if err != nil {
		// 	ctx.Log().Error("gel.RPC().GetRPC().Swap:", err)
		// }
		ctx.Log().Debug(string(body.GetResult()))
		ctx.Log().Debug(body.GetHeader())
		ctx.Log().Debug(body.GetStatus())
		time.Sleep(time.Second)
		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})
	return apiSrv
}

func mqcserver() transport.Server {
	mqcSrv := mqc.New("mqcserver")
	mqcSrv.Use(tracing.Server(tracing.WithPropagator(propagation.TraceContext{}), tracing.WithTracerProvider(otel.GetTracerProvider())))
	mqcSrv.Handle("/demomqc", func(ctx context.Context) interface{} {
		ctx.Log().Debug("demomqc")
		body, err := gel.Http().GetHttp().Swap(ctx, "http://apiserver/log/info", xhttp.WithMethod(http.MethodPost))
		if err != nil {
			ctx.Log().Error("gel.Http().GetHttp().Swap:", err)
		}
		ctx.Log().Debug(string(body.GetResult()))
		ctx.Log().Debug(body.GetHeader())
		ctx.Log().Debug(body.GetStatus())
		time.Sleep(time.Second * 2)
		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})

	return mqcSrv
}

func rpcserver() transport.Server {
	rpcSrv := rpc.New("rpcserver")
	rpcSrv.Use(tracing.Server(tracing.WithPropagator(propagation.TraceContext{}), tracing.WithTracerProvider(otel.GetTracerProvider())))
	rpcSrv.Handle("/demorpc", func(ctx context.Context) interface{} {
		time.Sleep(time.Second * 1)
		ctx.Log().Debug("demorpc")
		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})

	return rpcSrv
}

func cronserver() transport.Server {
	cronSrv := cron.New("cronserver")

	cronSrv.Handle("/democron", func(ctx context.Context) interface{} {
		ctx.Log().Debug("democron")

		gel.Queue().GetQueue("default").Send(ctx, "xx.xx.xx", map[string]interface{}{
			"a": time.Now().Unix(),
		})

		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})
	return cronSrv
}
