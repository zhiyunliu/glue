package defaultipc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"

	"context"

	ants "github.com/panjf2000/ants/v2"
	"github.com/zhiyunliu/alloter"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/metrics"
	"github.com/zhiyunliu/glue/xipc"
	"github.com/zhiyunliu/golibs/bytesconv"
)

// processor cron管理程序，用于管理多个任务的执行，暂停，恢复，动态添加，移除
type processor struct {
	endpoint   string
	engine     *alloter.Engine
	listener   net.Listener
	capactiy   int
	runnerPool *ants.PoolWithFunc
	onceLock   sync.Once
	closed     chan struct{}
	Metrics    *Metrics
}

// NewProcessor 创建processor
func newProcessor(engine *alloter.Engine, capactiy int, endpoint string) (s *processor, err error) {
	if capactiy <= 0 {
		capactiy = 100
	}

	s = &processor{
		engine:   engine,
		capactiy: capactiy,
		endpoint: endpoint,
	}

	s.listener, err = Listen(endpoint)
	if err != nil {
		return nil, err
	}
	s.Metrics = metrics.GetMetrics[Metrics](xipc.ScopeName+"-default", xipc.ScopeName)
	s.runnerPool, err = ants.NewPoolWithFunc(capactiy, func(data any) {
		s.Metrics.IpcProcessing.Add(context.Background(), 1)
		s.serveConn(data.(net.Conn))
		s.Metrics.IpcProcessing.Add(context.Background(), -1)
	}, ants.WithLogger(&loggerWrap{logger: log.DefaultLogger}))

	return s, err
}

// Start 所有任务
func (s *processor) Start() error {

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
		if err := s.runnerPool.Invoke(conn); err != nil {
			log.Errorf("ipc.runnerPool.Invoke,err:%+v", err)
		}
	}

}

// nolint
func (s *processor) serveConn(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			s.recordCloseError(fmt.Errorf("ipc: close client connection: %w", err))
		}
	}()

	for {
		payload, err := ReadFrame(conn)
		if err != nil {
			if !errors.Is(err, io.EOF) && !s.isClosed() {
				s.recordCloseError(fmt.Errorf("ipc: read request frame: %w", err))
			}
			return
		}
		var wire xipc.Request
		if err := json.Unmarshal(payload, &wire); err != nil {
			if err := s.writeResponse(conn, xipc.BadRequest(err)); err != nil {
				return
			}
			continue
		}
		ctx := context.Background()
		response := s.dispatch(ctx, &wire)
		if err := s.writeResponse(conn, response); err != nil {
			if !s.isClosed() {
				s.recordCloseError(err)
			}
			return
		}
	}
}

func (s *processor) isClosed() bool {
	select {
	case <-s.closed:
		return true
	default:
		return false
	}
}

func (s *processor) dispatch(ctx context.Context, request *xipc.Request) (response *xipc.Response) {
	response = &xipc.Response{}
	//转换输入参数

	req := newServerRequest(ctx, request)
	resp := newServerResponse()

	//发起本地处理
	err := s.engine.HandleRequest(req, resp)
	if err != nil {
		response.Status = int32(http.StatusInternalServerError)
		//nolint
		response.Result = bytesconv.StringToBytes(fmt.Sprintf("处理请求有误%s", err.Error()))
		return response
	}

	//处理响应内容
	//nolint
	response.Status = int32(resp.Status())
	response.Result = resp.buffer.Bytes()
	response.Headers = resp.Header()
	return response

}

func (s *processor) writeResponse(conn net.Conn, response *xipc.Response) error {
	payload, err := json.Marshal(response)
	if err != nil {

		return fmt.Errorf("ipc: encode fallback response: %w", err)
	}
	if err := WriteFrame(conn, payload); err != nil {
		return fmt.Errorf("ipc: write response frame: %w", err)
	}
	return nil
}

func (s *processor) recordCloseError(err error) {
	log.Error(err)
}

func (s *processor) Stop() error {
	var err error
	s.onceLock.Do(func() {
		close(s.closed)
		err = s.listener.Close()
	})
	return err
}
