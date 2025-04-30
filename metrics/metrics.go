package metrics

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/metric"
)

var (
	counterPtrType        = reflect.TypeOf((*metric.Int64Counter)(nil)).Elem()
	floatcounterPtrType   = reflect.TypeOf((*metric.Float64Counter)(nil)).Elem()
	gaugePtrType          = reflect.TypeOf((*metric.Int64Gauge)(nil)).Elem()
	floatGaugePtrType     = reflect.TypeOf((*metric.Float64Gauge)(nil)).Elem()
	timerPtrType          = reflect.TypeOf((*Timer)(nil)).Elem()
	histogramPtrType      = reflect.TypeOf((*metric.Int64Histogram)(nil)).Elem()
	floatHistogramPtrType = reflect.TypeOf((*metric.Float64Histogram)(nil)).Elem()
)

// Init initializes the metrics with the given factory and config.
func Init(m any, factory *Factory) error {
	if m == nil {
		return fmt.Errorf("metrics.Init: m cannot be nil")

	}

	rv := reflect.ValueOf(m)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("metrics.Init: m must be a pointer to a struct")
	}

	v := rv.Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		metricName := field.Tag.Get("metric")
		if metricName == "" {
			//没有配置metric标签，跳过
			continue
		}

		opts, err := prepareOptions(metricName, &field)
		if err != nil {
			return err
		}

		descOpt := metric.WithDescription(opts.Help)

		var obj any
		switch {
		case field.Type.AssignableTo(counterPtrType):
			obj, err = factory.CreateIntCounter(metricName, descOpt)
		case field.Type.AssignableTo(floatcounterPtrType):
			obj, err = factory.CreateFloatCounter(metricName, descOpt)

		case field.Type.AssignableTo(gaugePtrType):
			obj, err = factory.CreateIntGauge(metricName, descOpt)
		case field.Type.AssignableTo(floatGaugePtrType):
			obj, err = factory.CreateFloatGauge(metricName, descOpt)

		case field.Type.AssignableTo(histogramPtrType):
			obj, err = factory.CreateIntHistogram(metricName, descOpt, metric.WithExplicitBucketBoundaries(opts.Buckets...))
		case field.Type.AssignableTo(floatHistogramPtrType):
			obj, err = factory.CreateFloatHistogram(metricName, descOpt, metric.WithExplicitBucketBoundaries(opts.Buckets...))

		case field.Type.AssignableTo(timerPtrType):
			obj, err = factory.CreateTimer(metricName, descOpt, metric.WithExplicitBucketBoundaries(opts.Buckets...))

		default:
			continue
		}
		if err != nil {
			return err
		}

		v.Field(i).Set(reflect.ValueOf(obj))
	}
	return nil
}

func prepareOptions(metricName string, field *reflect.StructField) (*Options, error) {

	tagVal := field.Tag.Get("lbls")
	if tagVal == "" {
		return nil, fmt.Errorf("metrics.Init:lbls tag is required for metric %s", metricName)
	}

	lbls := strings.Split(tagVal, ",")
	opts := []Option{WithName(metricName), WithLabels(lbls)}

	buckets := field.Tag.Get("buckets")
	if buckets != "" {
		bks := strings.Split(buckets, ",")
		bksv := make([]float64, len(bks))
		for i, b := range bks {
			bksv[i], _ = strconv.ParseFloat(b, 64)
		}
		opts = append(opts, WithBuckets(bksv))
	}

	mopts := &Options{}
	for _, opt := range opts {
		opt(mopts)
	}

	return mopts, nil
}
