package opentelemetry

// import (
// 	stdmetrics "github.com/zhiyunliu/glue/metrics"
// 	"github.com/zhiyunliu/glue/opentelemetry/metrics"
// 	"github.com/zhiyunliu/glue/standard"
// )

// // GetMetricsObserver returns a metrics observer for the specified protocol name.
// func GetMetricsObserver(protoName string) *metrics.Observer {
// 	stdInstance := standard.GetInstance(stdmetrics.TypeNode).(stdmetrics.StandardMetric)
// 	metricsProvider := stdInstance.GetProvider(protoName)
// 	metricsObserver := metrics.NewObserver(metricsProvider)
// 	return metricsObserver
// }
