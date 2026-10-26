package ipc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
)

// Client 在单条连接上执行同步命令调用。
type Client struct {
	conn net.Conn
	mu   sync.Mutex

	closed   bool
	closeErr error
}

// Dial 连接到指定 IPC 端点。
func Dial(endpoint string) (*Client, error) {
	conn, err := dialTransport(endpoint)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}

// Call 发送命令并等待响应。同一 Client 上的调用会串行执行。
func (c *Client) Call(command string, data any) (*Response, error) {
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
	var response Response
	if err := json.Unmarshal(responsePayload, &response); err != nil {
		return nil, c.closeAfterCallError(fmt.Errorf("ipc: decode response: %w", err))
	}
	return &response, nil
}

func (c *Client) closeAfterCallError(callErr error) error {
	c.closed = true
	if err := c.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return errors.Join(callErr, fmt.Errorf("ipc: close broken connection: %w", err))
	}
	return callErr
}

// Close 关闭客户端连接。
func (c *Client) Close() error {
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
