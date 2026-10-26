package ipc

import (
	"context"
	"errors"
	"net"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/transport"
)

// Server 将本地命令通道接入框架生命周期和统一路由。
type Server struct {
	closed   bool
	name     string
	opts     *options
	mu       sync.Mutex
	handlers map[string]Handler
	command  *commandServer
	listener net.Listener
	done     chan struct{}
	stopDone chan struct{}
	stopErr  error
	serveErr error
	cancel   context.CancelFunc
}

var _ transport.Server = (*Server)(nil)
var _ transport.Endpointer = (*Server)(nil)

// New 创建服务但不监听端点，路由和参数应在启动前配置。
func New(name string, opts ...Option) *Server {
	s := &Server{name: name, opts: setDefaultOption(), handlers: make(map[string]Handler)}
	s.Options(opts...)
	return s
}

// NewServer 保留创建时监听的低层接口，需要调用 Serve 接受连接。
func NewServer(endpoint string) (*Server, error) {
	s := New("", WithAddr(endpoint))
	c, err := newCommandServer(endpoint)
	if err != nil {
		return nil, err
	}
	s.command = c
	s.listener = c.listener
	return s, nil
}

// Options 设置启动前参数。
func (s *Server) Options(opts ...Option) {
	for _, opt := range opts {
		opt(s.opts)
	}
}
func (s *Server) Name() string {
	if s.name == "" {
		return "ipc"
	}
	return s.name
}
func (s *Server) Type() string { return "ipc" }
func (s *Server) Config(cfg config.Config) error {
	if cfg == nil {
		return nil
	}
	s.opts.config = cfg
	return cfg.Get("servers." + s.Name()).ScanTo(s.opts.srvCfg)
}
func (s *Server) ServiceName() string { return s.opts.serviceName }

// Address 返回 Dial 所需的原始本地地址。
func (s *Server) Address() string { return s.opts.srvCfg.Config.Addr }

// Endpoint 返回注册用 URL，原始地址保存在路径中，兼容管道名及绝对路径。
func (s *Server) Endpoint() *url.URL { return &url.URL{Scheme: "ipc", Path: s.Address()} }
func (s *Server) StopMaximumTimeout() time.Duration {
	seconds := s.opts.srvCfg.Config.StopMaximumTimeout
	if seconds == 0 {
		seconds = 15
	}
	return time.Duration(seconds) * time.Second
}
func (s *Server) Attempt() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.done == nil }
func (s *Server) Use(mids ...middleware.Middleware) {
	s.opts.middlewares = append(s.opts.middlewares, mids...)
}
func (s *Server) Group(path string, mids ...middleware.Middleware) *engine.RouterGroup {
	return s.opts.router.Group(path, mids...)
}
func (s *Server) Handle(path string, obj interface{}, opts ...engine.RouterOption) {
	s.opts.router.Handle(path, obj, opts...)
}
func (s *Server) RouterPathList() transport.RouterList {
	paths := s.opts.router.GetTreePathList()
	s.mu.Lock()
	for name := range s.handlers {
		paths = append(paths, name)
	}
	s.mu.Unlock()
	sort.Strings(paths)
	return engine.RouterList{ServerType: s.Type(), PathList: paths}
}

// RegisterCommand 保留低层命令优先匹配，支持启动后注册。
func (s *Server) RegisterCommand(name string, h Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[name] = h
	if s.command != nil {
		s.command.RegisterCommand(name, h)
	}
}
func (s *Server) RegisterFunc(name string, h HandlerFunc) { s.RegisterCommand(name, h) }

func (s *Server) launch(ctx context.Context) error {
	if s.closed {
		return net.ErrClosed
	}
	if s.done != nil {
		return ErrServerAlreadyServing
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(transport.WithServerContext(ctx, s))
	fallback, err := s.resolveEngineRoute(runCtx)
	if err != nil {
		cancel()
		return err
	}
	if s.command == nil {
		s.command, err = newCommandServer(s.Address())
		if err != nil {
			cancel()
			return err
		}
	}
	c := s.command
	c.fallback = fallback
	for name, h := range s.handlers {
		c.RegisterCommand(name, h)
	}
	s.listener = c.listener
	s.cancel = cancel
	s.serveErr = nil
	s.stopDone = nil
	s.stopErr = nil
	s.done = make(chan struct{})
	done := s.done
	go func() { err := c.Serve(); s.mu.Lock(); s.serveErr = err; s.mu.Unlock(); close(done) }()
	return nil
}

// Start 绑定端点并启动后台接收，重复启动返回明确错误。
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.opts.srvCfg.Config.Status == engine.StatusStop {
		s.mu.Unlock()
		return nil
	}
	err := s.launch(ctx)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	for _, hook := range s.opts.startedHooks {
		if err := hook(ctx); err != nil {
			return errors.Join(err, s.Stop(context.Background()))
		}
	}
	return nil
}

// Serve 保留阻塞接口，与 Start 互斥。
func (s *Server) Serve() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	if s.opts.srvCfg.Config.Status == engine.StatusStop {
		s.mu.Unlock()
		return nil
	}
	err := s.launch(context.Background())
	done := s.done
	s.mu.Unlock()
	if err != nil {
		return err
	}
	<-done
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.serveErr
}

// Stop 立即关闭监听器和连接，再在上下文及最大时限内等待处理器退出。
func (s *Server) Stop(ctx context.Context) error {
	timeout, cancel := context.WithTimeout(ctx, s.StopMaximumTimeout())
	defer cancel()
	s.mu.Lock()
	if s.command == nil {
		s.mu.Unlock()
		return nil
	}
	if s.stopDone == nil {
		s.stopDone = make(chan struct{})
		c, done, stopped := s.command, s.done, s.stopDone
		if s.cancel != nil {
			s.cancel()
		}
		go func() {
			err := c.Close()
			if done != nil {
				<-done
			}
			for _, hook := range s.opts.endHooks {
				err = errors.Join(err, hook(timeout))
			}
			s.mu.Lock()
			s.stopErr = errors.Join(err, s.serveErr)
			s.command = nil
			s.listener = nil
			s.done = nil
			s.mu.Unlock()
			close(stopped)
		}()
	}
	stopped := s.stopDone
	s.mu.Unlock()
	select {
	case <-stopped:
		s.mu.Lock()
		err := s.stopErr
		s.mu.Unlock()
		return err
	case <-timeout.Done():
		return timeout.Err()
	}
}

// Close 使用默认停止上限关闭服务。
func (s *Server) Close() error {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	return s.Stop(context.Background())
}
