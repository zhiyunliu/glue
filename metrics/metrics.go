package metrics

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/zhiyunliu/glue/config"
)

var (
	counterPtrType   = reflect.TypeOf((*Counter)(nil)).Elem()
	gaugePtrType     = reflect.TypeOf((*Gauge)(nil)).Elem()
	timerPtrType     = reflect.TypeOf((*Timer)(nil)).Elem()
	histogramPtrType = reflect.TypeOf((*Histogram)(nil)).Elem()
)

// Init initializes the metrics with the given factory and config.
func Init(m any, factory Factory, config config.Config) error {
	if factory == nil {
		factory = noopFactory
	}

	rv := reflect.ValueOf(m)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("metrics.Init: m must be a pointer to a struct")
	}

	v := rv.Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		metric := field.Tag.Get("metric")
		if metric == "" {
			//没有配置metric标签，跳过
			continue
		}

		opts, err := prepareOptions(metric, &field)
		if err != nil {
			return err
		}

		mcfg := config.Get(metric)
		var obj any
		switch {
		case field.Type.AssignableTo(counterPtrType):
			obj = factory.Counter(mcfg, opts)
		case field.Type.AssignableTo(gaugePtrType):
			obj = factory.Gauge(mcfg, opts)
		case field.Type.AssignableTo(timerPtrType):
			obj = factory.Timer(mcfg, opts)
		case field.Type.AssignableTo(histogramPtrType):
			obj = factory.Histogram(mcfg, opts)
		default:
			continue
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

	namespace := field.Tag.Get("namespace")
	if namespace != "" {
		opts = append(opts, WithNamespace(namespace))
	}

	subsystem := field.Tag.Get("subsystem")
	if subsystem != "" {
		opts = append(opts, WithSubsystem(subsystem))
	}
	mopts := &Options{}
	for _, opt := range opts {
		opt(mopts)
	}

	return mopts, nil
}
