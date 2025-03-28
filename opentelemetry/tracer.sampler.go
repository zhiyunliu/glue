package opentelemetry

import (
	"math/rand/v2"
	"sync/atomic"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	zeroTraceID trace.TraceID
	_           sdktrace.Sampler = (*DynamicSampler)(nil)
)

// 1. 实现动态采样器
type DynamicSampler struct {
	//0~100
	rateBits uint64 // 使用原子操作存储的采样率
}

func NewDynamicSampler(initialRate uint64) sdktrace.Sampler {
	return &DynamicSampler{
		rateBits: initialRate,
	}
}

func (s *DynamicSampler) SetRate(rate uint64) {
	atomic.StoreUint64(&s.rateBits, rate)
}

func (s *DynamicSampler) GetRate() uint64 {
	return atomic.LoadUint64(&s.rateBits)
}

func (s *DynamicSampler) ShouldSample(p sdktrace.SamplingParameters) sdktrace.SamplingResult {
	rate := s.GetRate()
	if rate >= 1 {
		return sdktrace.SamplingResult{Decision: sdktrace.RecordAndSample}
	}
	//todo:这个地方需要检查远程的Span采样情况

	if rate <= 0 || rand.Uint64N(100) > rate {
		return sdktrace.SamplingResult{Decision: sdktrace.Drop}
	}

	return sdktrace.SamplingResult{Decision: sdktrace.RecordAndSample}
}

func (s *DynamicSampler) Description() string {
	return "DynamicSampler"
}
