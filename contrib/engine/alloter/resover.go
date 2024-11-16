package alloter

import (
	"github.com/zhiyunliu/alloter"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
)

func init() {
	engine.Register(&alloterEngineResover{})
}

type alloterEngineResover struct{}

func (r *alloterEngineResover) Name() string {
	return "alloter"
}
func (r *alloterEngineResover) Resolve(name string, config config.Config, opts ...engine.Option) (engine.AdapterEngine, error) {
	innerEngine := alloter.New()
	return NewAlloterEngine(innerEngine, opts...), nil
}
