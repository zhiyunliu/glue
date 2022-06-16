package main

import (
	"fmt"

	gel "github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/glue/context"
	_ "github.com/zhiyunliu/glue/contrib/cache/redis"
	_ "github.com/zhiyunliu/glue/contrib/config/consul"
	_ "github.com/zhiyunliu/glue/contrib/config/nacos"
	_ "github.com/zhiyunliu/glue/contrib/queue/redis"
	_ "github.com/zhiyunliu/glue/contrib/registry/nacos"
	_ "github.com/zhiyunliu/glue/contrib/xdb/mysql"

	//_ "github.com/zhiyunliu/glue/contrib/xdb/oracle"
	_ "github.com/zhiyunliu/glue/contrib/xdb/postgres"
	_ "github.com/zhiyunliu/glue/contrib/xdb/sqlite"
	_ "github.com/zhiyunliu/glue/contrib/xdb/sqlserver"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/zhiyunliu/glue/middleware/ratelimit"
	"github.com/zhiyunliu/glue/middleware/tracing"

	_ "github.com/zhiyunliu/glue/contrib/dlocker/redis"

	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/examples/apiserver/demos"
	"github.com/zhiyunliu/glue/server/api"
	"github.com/zhiyunliu/golibs/xtypes"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var Name = "apiserver"

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

func main() {
	setTracerProvider("http://127.0.0.1:14268/api/traces")

	apiSrv := api.New("apiserver")
	//mqcSrv := mqc.New("bb")

	apiSrv.Handle("/demo", func(ctx context.Context) interface{} {
		ctx.Log().Debug("demo")
		return xtypes.XMap{
			"a": 1,
			"b": 2,
		}
	})

	apiSrv.Handle("/error", func(ctx context.Context) interface{} {
		ctx.Log().Debug("error")
		return errors.New(300, "xxx")
	})

	apiSrv.Handle("/panic", func(ctx context.Context) interface{} {
		ctx.Log().Debug("panic")
		panic(fmt.Errorf("xx i am panic"))
	})

	apiSrv.Handle("/db", demos.NewDb())
	apiSrv.Handle("/cache", demos.NewCache())
	apiSrv.Handle("/queue", demos.NewQueue())
	apiSrv.Handle("/log", demos.NewLogDemo())
	apiSrv.Handle("/rpc", demos.NewGrpcDemo())

	//apiSrv.Use(jwt.Server(jwt.WithSecret("123456")))
	apiSrv.Use(ratelimit.Server())
	//apiSrv.Use(tracing.Server(tracing.WithTracerProvider(provider)))
	apiSrv.Use(tracing.Server(tracing.WithPropagator(propagation.TraceContext{}), tracing.WithTracerProvider(otel.GetTracerProvider())))

	app := gel.NewApp(gel.Server(apiSrv), gel.LogConcurrency(1))
	app.Start()
}
