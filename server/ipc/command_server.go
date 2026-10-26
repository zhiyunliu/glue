package ipc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
)

type wireRequest struct {
	Cmd  string          `json:"cmd"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Server 接收连接并分发已注册的命令。
type commandServer struct {
	endpoint string
	listener net.Listener

	handlerMu sync.RWMutex
	handlers  map[string]Handler

	connMu    sync.Mutex
	conns     map[net.Conn]struct{}
	closeErrs []error
	connWG    sync.WaitGroup

	serveMu sync.Mutex
	serving bool
	closed  chan struct{}

	closeOnce sync.Once
	closeErr  error
	fallback  func(Request) *Response
}

// NewServer 创建并监听指定端点。
func newCommandServer(endpoint string) (*commandServer, error) {
	listener, err := Listen(endpoint)
	if err != nil {
		return nil, err
	}
	return &commandServer{
		endpoint: endpoint,
		listener: listener,
		handlers: make(map[string]Handler),
		conns:    make(map[net.Conn]struct{}),
		closed:   make(chan struct{}),
	}, nil
}

// RegisterCommand 注册命令处理器。
func (s *commandServer) RegisterCommand(name string, handler Handler) {
	s.handlerMu.Lock()
	s.handlers[name] = handler
	s.handlerMu.Unlock()
}

// RegisterFunc 以函数形式注册命令处理器。
func (s *commandServer) RegisterFunc(name string, handler HandlerFunc) {
	s.RegisterCommand(name, handler)
}

// Serve 阻塞接受连接，直到服务关闭或发生监听错误。
func (s *commandServer) Serve() error {
	s.serveMu.Lock()
	if s.serving {
		s.serveMu.Unlock()
		return ErrServerAlreadyServing
	}
	s.serving = true
	s.serveMu.Unlock()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.closed:
				return nil
			default:
				return fmt.Errorf("ipc: accept connection: %w", err)
			}
		}

		s.connMu.Lock()
		select {
		case <-s.closed:
			s.connMu.Unlock()
			if err := conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				s.recordCloseError(fmt.Errorf("ipc: close rejected connection: %w", err))
			}
			return nil
		default:
			s.conns[conn] = struct{}{}
			s.connWG.Add(1)
		}
		s.connMu.Unlock()

		go s.serveConn(conn)
	}
}

func (s *commandServer) serveConn(conn net.Conn) {
	defer s.connWG.Done()
	defer func() {
		if err := conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			s.recordCloseError(fmt.Errorf("ipc: close client connection: %w", err))
		}
		s.connMu.Lock()
		delete(s.conns, conn)
		s.connMu.Unlock()
	}()

	for {
		payload, err := readFrame(conn)
		if err != nil {
			if !errors.Is(err, io.EOF) && !s.isClosed() {
				s.recordCloseError(fmt.Errorf("ipc: read request frame: %w", err))
			}
			return
		}
		var wire wireRequest
		if err := json.Unmarshal(payload, &wire); err != nil {
			if err := s.writeResponse(conn, Fail(CodeBadRequest, "bad_frame", err.Error())); err != nil {
				return
			}
			continue
		}

		response := s.dispatch(&jsonRequest{cmd: wire.Cmd, data: wire.Data})
		if err := s.writeResponse(conn, response); err != nil {
			if !s.isClosed() {
				s.recordCloseError(err)
			}
			return
		}
	}
}

func (s *commandServer) isClosed() bool {
	select {
	case <-s.closed:
		return true
	default:
		return false
	}
}

func (s *commandServer) dispatch(request Request) (response *Response) {
	s.handlerMu.RLock()
	handler, ok := s.handlers[request.Command()]
	s.handlerMu.RUnlock()
	if !ok {
		if s.fallback != nil {
			return s.fallback(request)
		}
		return Fail(CodeUnknownCommand, "unknown_command", "unknown command: "+request.Command())
	}
	if handler == nil {
		return Fail(CodeHandlerError, "nil_handler", "command handler is nil")
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			response = Fail(CodeHandlerPanic, "handler_panic", fmt.Sprint(recovered))
		}
	}()
	response = handler.Handle(request)
	if response == nil {
		return Fail(CodeHandlerError, "nil_response", "command handler returned a nil response")
	}
	return response
}

func (s *commandServer) writeResponse(conn net.Conn, response *Response) error {
	payload, err := json.Marshal(response)
	if err != nil {
		payload, err = json.Marshal(Fail(CodeEncodeError, "encode_error", err.Error()))
		if err != nil {
			return fmt.Errorf("ipc: encode fallback response: %w", err)
		}
	}
	if err := writeFrame(conn, payload); err != nil {
		return fmt.Errorf("ipc: write response frame: %w", err)
	}
	return nil
}

func (s *commandServer) recordCloseError(err error) {
	s.connMu.Lock()
	s.closeErrs = append(s.closeErrs, err)
	s.connMu.Unlock()
}

// Close 关闭监听器和活动连接，并等待连接处理协程退出。
func (s *commandServer) Close() error {
	s.closeOnce.Do(func() {
		close(s.closed)
		if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			s.closeErr = errors.Join(s.closeErr, fmt.Errorf("ipc: close listener: %w", err))
		}

		s.connMu.Lock()
		for conn := range s.conns {
			if err := conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				s.closeErrs = append(s.closeErrs, fmt.Errorf("ipc: close client connection: %w", err))
			}
		}
		s.connMu.Unlock()
		s.connWG.Wait()

		s.connMu.Lock()
		for _, err := range s.closeErrs {
			s.closeErr = errors.Join(s.closeErr, err)
		}
		s.connMu.Unlock()
	})
	return s.closeErr
}

// Address 返回服务端监听的原始端点。
func (s *commandServer) Address() string {
	return s.endpoint
}
