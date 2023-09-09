package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
)

func init() {
	engine.Register(&ginEngineResover{})
}

type ginEngineResover struct{}

func (r *ginEngineResover) Name() string {
	return "gin"
}
func (r *ginEngineResover) Resolve(name string, config config.Config, opts ...engine.Option) (engine.AdapterEngine, error) {
	gin.SetMode(gin.ReleaseMode)
	innerEngine := gin.New()
	return NewGinEngine(innerEngine, opts...), nil
}
