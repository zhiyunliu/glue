package defaultipc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/xipc"
)

// Client 在单条连接上执行同步命令调用。
type endpointer struct {
	conn net.Conn
	mu   sync.Mutex

	closed   bool
	closeErr error
}

type xipcClient struct {
	requests     cmap.ConcurrentMap[string, *endpointer]
	clientConfig config.Config
}

// NewClient 连接到指定 IPC 端点。
func NewClient(clientCfg config.Config) (Client, error) {
	req := &xipcClient{
		clientConfig: clientCfg,
		requests:     cmap.New[*endpointer](),
	}
	return req, nil

}

func (p *xipcClient) Request(ctx context.Context, path string, data any, opts ...RequestOption) (res Body, err error) {
	return
}

func (p *xipcClient) Call(ctx context.Context, srvName string, command string, data any, opts ...RequestOption) (res Body, err error) {
	return
}

func (p *xipcClient) getClient(endpoint string) (*endpointer, error) {
	conn, err := dialTransport(endpoint)
	if err != nil {
		return nil, err
	}
	return &endpointer{conn: conn}, nil
}

// Call 发送命令并等待响应。同一 Client 上的调用会串行执行。
func (c *endpointer) Call(command string, data any) (*xipc.Response, error) {
	var rawData json.RawMessage
	if data != nil {
		encoded, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("ipc: encode request data: %w", err)
		}
		rawData = encoded
	}
	payload, err := json.Marshal(wireRequest{Cmd: command, Data: rawData})
	if err != nil {
		return nil, fmt.Errorf("ipc: encode request: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, ErrClosed
	}
	if err := writeFrame(c.conn, payload); err != nil {
		return nil, c.closeAfterCallError(fmt.Errorf("ipc: write request: %w", err))
	}
	responsePayload, err := readFrame(c.conn)
	if err != nil {
		return nil, c.closeAfterCallError(fmt.Errorf("ipc: read response: %w", err))
	}
	var response xipc.Response
	if err := json.Unmarshal(responsePayload, &response); err != nil {
		return nil, c.closeAfterCallError(fmt.Errorf("ipc: decode response: %w", err))
	}
	return &response, nil
}

func (c *endpointer) closeAfterCallError(callErr error) error {
	c.closed = true
	if err := c.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return errors.Join(callErr, fmt.Errorf("ipc: close broken connection: %w", err))
	}
	return callErr
}

// Close 关闭客户端连接。
func (c *endpointer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return c.closeErr
	}
	c.closed = true
	c.closeErr = c.conn.Close()
	if errors.Is(c.closeErr, net.ErrClosed) {
		c.closeErr = nil
		return nil
	}
	if c.closeErr != nil {
		return fmt.Errorf("ipc: close connection: %w", c.closeErr)
	}
	return nil
}

// Client 定义同步命令客户端及其资源释放契约。
type Client = xipc.Client
type RequestOption = xipc.RequestOption
type Body = xipc.Body
type ResponseEntity = xipc.ResponseEntity

type wireRequest struct {
	Cmd  string          `json:"cmd"`
	Data json.RawMessage `json:"data,omitempty"`
}
