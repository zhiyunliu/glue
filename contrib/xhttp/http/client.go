package http

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/zhiyunliu/glue/constants"
	"github.com/zhiyunliu/glue/contrib/xhttp/http/balancer"
	"github.com/zhiyunliu/glue/middleware/tracing"
	"github.com/zhiyunliu/glue/registry"
	"github.com/zhiyunliu/glue/selector"
	"github.com/zhiyunliu/glue/xhttp"
	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/httputil"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	registrar registry.Registrar
	setting   *setting
	client    *http.Client
	selector  selector.Selector
	ctx       context.Context
	ctxCancel context.CancelFunc
	tracer    *tracing.Tracer
}

// NewClientByConf 创建RPC客户端,地址是远程RPC服务器地址或注册中心地址
func NewClient(registrar registry.Registrar, setting *setting, reqPath *url.URL) (*Client, error) {
	client := &Client{
		registrar: registrar,
		setting:   setting,
		client:    &http.Client{},
	}

	tlsCfg, err := client.getTlsConfig()
	if err != nil {
		return nil, err
	}
	if setting.Trace {
		client.tracer = tracing.NewTracer(trace.SpanKindClient)
	}
	client.ctx, client.ctxCancel = context.WithCancel(context.Background())

	client.selector, err = balancer.NewSelector(client.ctx, registrar, reqPath, setting.Balancer)
	if err != nil {
		return nil, err
	}

	client.client.Transport = &http.Transport{
		TLSClientConfig: tlsCfg,
		Proxy:           client.getProxy(),
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(setting.ConnTimeout) * time.Second,
			KeepAlive: time.Duration(setting.KeepaliveTimeout) * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          setting.MaxIdleConns,
		IdleConnTimeout:       time.Duration(setting.IdleConnTimeout) * time.Second,
		TLSHandshakeTimeout:   time.Duration(setting.TLSHandshakeTimeout) * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return client, nil
}

// RequestByString 发送Request请求
func (c *Client) RequestByString(ctx context.Context, reqPath *url.URL, input any, opt *xhttp.Options) (res xhttp.Body, err error) {
	//处理可选参数

	var bodyBytes []byte

	switch t := input.(type) {
	case []byte:
		bodyBytes = t
	case string:
		bodyBytes = bytesconv.StringToBytes(t)
	case *string:
		bodyBytes = bytesconv.StringToBytes(*t)
	default:
		bodyBytes, _ = json.Marshal(t)
		xhttp.WithContentType(constants.ContentTypeApplicationJSON)(opt)
	}

	response, err := c.clientRequest(ctx, reqPath, opt, bodyBytes)
	if err != nil {
		return newBodyByError(err), err
	}
	return response, err
}

// Close 关闭RPC客户端连接
func (c *Client) Close() {
	if c.ctxCancel != nil {
		c.ctxCancel()
	}
}

func (c *Client) clientRequest(ctx context.Context, reqPath *url.URL, o *xhttp.Options, input []byte) (response xhttp.Body, err error) {
	node, err := c.getServiceNode(ctx)
	if err != nil {
		return nil, err
	}

	httpOpts := make([]httputil.Option, 0)
	for k, v := range o.Header {
		httpOpts = append(httpOpts, httputil.WithHeader(k, v))
	}
	httpOpts = append(httpOpts, httputil.WithClient(c.client))
	if o.Handler != nil {
		httpOpts = append(httpOpts, httputil.WithRespHandler(o.Handler))
	}

	queryParam := ""
	if reqPath.RawQuery != "" {
		queryParam = "?" + reqPath.RawQuery
	}
	return httputil.Request(o.Method, fmt.Sprintf("%s%s%s", node.Address(), reqPath.Path, queryParam), input, httpOpts...)
}

func (c *Client) getServiceNode(ctx context.Context) (selector.Node, error) {
	node, done, err := c.selector.Select(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		done(ctx, selector.DoneInfo{Err: err})
	}()
	return node, nil
}

func (c *Client) getTlsConfig() (*tls.Config, error) {
	ssl := &tls.Config{InsecureSkipVerify: true}
	if c.setting.CertFile != "" && c.setting.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(c.setting.CertFile, c.setting.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("LoadX509KeyPair(CertFile: %s, KeyFile: %s),error:%v", c.setting.CertFile, c.setting.KeyFile, err)
		}
		ssl.Certificates = []tls.Certificate{cert}
	}
	if c.setting.CaFile != "" {
		caData, err := os.ReadFile(c.setting.CaFile)
		if err != nil {
			return nil, fmt.Errorf("CaFile(%s) error:%v", c.setting.CaFile, err)
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caData)
		ssl.RootCAs = pool
	}
	if len(ssl.Certificates) == 0 && ssl.RootCAs == nil {
		return ssl, nil
	}
	ssl.Rand = rand.Reader
	return ssl, nil
}

func (c *Client) getProxy() func(*http.Request) (*url.URL, error) {
	if c.setting.ProxyURL != "" {
		proxyURL, err := url.Parse(c.setting.ProxyURL)
		return func(_ *http.Request) (*url.URL, error) {
			return proxyURL, err
		}
	}
	return nil
}
