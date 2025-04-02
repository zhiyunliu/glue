package opentelemetry

import (
	"math/rand/v2"
	"sync/atomic"

	"github.com/zhiyunliu/glue/config"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var (
	_ sdktrace.Sampler = (*DynamicSampler)(nil)
)

const (
	SamplerMaxRate = 100
)

// 1. 实现动态采样器
type DynamicSampler struct {
	//0~100
	rateBits    int64 // 使用原子操作存储的采样率
	serviceName string
	config      config.Config
}

func NewDynamicSampler(serviceName string, initialRate int64, config config.Config) *DynamicSampler {
	return &DynamicSampler{
		rateBits:    initialRate,
		serviceName: serviceName,
		config:      config,
	}
}

func (s *DynamicSampler) SetRate(rate int64) {
	atomic.StoreInt64(&s.rateBits, rate)
}

func (s *DynamicSampler) GetRate() int64 {
	return atomic.LoadInt64(&s.rateBits)
}

func (s *DynamicSampler) ShouldSample(p sdktrace.SamplingParameters) sdktrace.SamplingResult {
	rate := s.GetRate()
	if rate >= SamplerMaxRate {
		return sdktrace.SamplingResult{Decision: sdktrace.RecordAndSample}
	}
	if rand.Int64N(SamplerMaxRate) > rate {
		return sdktrace.SamplingResult{Decision: sdktrace.Drop}
	}

	return sdktrace.SamplingResult{Decision: sdktrace.RecordAndSample}
}

func (s *DynamicSampler) Description() string {
	return "DynamicSampler"
}

func (s *DynamicSampler) Watch() error {
	return s.config.Watch("sampler_rate", func(key string, val config.Value) {
		rateVal, _ := val.Int()
		s.SetRate(rateVal)
	})
}
