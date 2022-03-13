package mqc

import (
	"context"

	"github.com/zhiyunliu/velocity/transport"
)

type Server struct {
	name    string
	ctx     context.Context
	opts    options
	started bool
}

var _ transport.Server = (*Server)(nil)

// New 实例化
func New(name string, opts ...Option) *Server {
	s := &Server{
		name: name,
		opts: setDefaultOption(),
	}
	s.Options(opts...)
	return s
}

// Options 设置参数
func (e *Server) Options(opts ...Option) {
	for _, o := range opts {
		o(&e.opts)
	}
}

func (e *Server) Name() string {
	return e.name
}

// Start 开始
func (e *Server) Start(ctx context.Context) error {
	return nil
}

// Attempt 判断是否可以启动
func (e *Server) Attempt() bool {
	return !e.started
}

// Shutdown 停止
func (e *Server) Stop(ctx context.Context) error {
	return nil
}
