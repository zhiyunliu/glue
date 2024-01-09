package robfigcron

import (
	"fmt"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/xcron"
)

type serverResolver struct {
}

func (s *serverResolver) Name() string {
	return Proto
}

func (s *serverResolver) Resolve(name string,
	router *engine.RouterGroup,
	cfg config.Config,
	opts ...engine.Option) (xcron.Server, error) {
	setval := &serverConfig{}
	err := cfg.Scan(setval)
	if err != nil {
		return nil, fmt.Errorf("读取xcron配置[%s]错误%w", cfg.Path(), err)
	}

	return newServer(setval, router, opts...)
}

func init() {
	xcron.RegisterServer(&serverResolver{})
}
