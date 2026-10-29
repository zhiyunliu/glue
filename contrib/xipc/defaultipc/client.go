package defaultipc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"

	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/xipc"
	"github.com/zhiyunliu/golibs/bytesconv"
)

// Client 在单条连接上执行同步命令调用。
type endpointer struct {
	srvName  string
	conn     net.Conn
	mu       sync.Mutex
	client   *xipcClient
	closed   bool
	closeErr error
}

type xipcClient struct {
	requests     cmap.ConcurrentMap[string, *endpointer]
	clientConfig config.Config
	mutex        sync.Mutex
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
	urlVal, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("ipc.request path[%s]格式不符", path)
	}
	srvName := urlVal.Hostname()
	return p.Call(ctx, srvName, urlVal.Path, data, opts...)
}

func (p *xipcClient) Call(ctx context.Context, srvName string, command string, data any, opts ...RequestOption) (res Body, err error) {
	srvName = xipc.GetIpcName(srvName)

	endpointer, err := p.getEndpointer(srvName)
	if err != nil {
		return
	}
	//处理可选参数
	ipcOpt := &xipc.Options{
		Method: http.MethodPost,
		Header: make(map[string]string),
	}
	for i := range opts {
		opts[i](ipcOpt)
	}
	ipcResp, err := endpointer.Call(ctx, command, data, ipcOpt)
	return ipcResp, err
}

func (p *xipcClient) getEndpointer(srvName string) (*endpointer, error) {
	ep, ok := p.requests.Get(srvName)
	if ok {
		return ep, nil
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	conn, err := dialTransport(srvName)
	if err != nil {
		return nil, err
	}
	return &endpointer{
		srvName: srvName,
		conn:    conn,
		client:  p,
	}, nil
}

// Call 发送命令并等待响应。同一 Client 上的调用会串行执行。
func (c *endpointer) Call(ctx context.Context, command string, data any, ipcOpt *xipc.Options) (*xipc.Response, error) {

	var rawData []byte
	if data != nil {
		switch t := data.(type) {
		case []byte:
			rawData = t
		case string:
			rawData = bytesconv.StringToBytes(t)
		case *string:
			rawData = bytesconv.StringToBytes(*t)
		default:
			rawData, _ = json.Marshal(data)
			xipc.WithContentType(constants.ContentTypeApplicationJSON)(ipcOpt)
		}
	}
	if len(ipcOpt.Query) > 0 {
		command = fmt.Sprintf("%s?%s", command, ipcOpt.Query)
	}

	payload, err := json.Marshal(xipc.Request{
		Service: command,
		Method:  ipcOpt.Method,
		Header:  ipcOpt.Header,
		Body:    rawData,
	})
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
	if err := c.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
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
	c.client.requests.Remove(c.srvName)
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
