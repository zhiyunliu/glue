package xrpc

import (
	sctx "context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/context"
)

type Client interface {
	//Swap 将当前请求参数作为RPC参数并发送RPC请求
	Swap(ctx context.Context, service string, opts ...RequestOption) (res Body, err error)

	//RequestByCtx RPC请求，可通过context撤销请求
	Request(ctx sctx.Context, service string, input interface{}, opts ...RequestOption) (res Body, err error)

	//StreamRequest 发送流式RPC请求
	StreamRequest(ctx sctx.Context, service string, processor StreamProcessor, opts ...RequestOption) (err error)
}

// ClientResover 定义配置文件转换方法
type ClientResover interface {
	Name() string
	Resolve(name string, setting config.Config) (Client, error)
}

var clientResolvers = make(map[string]ClientResover)

// Register 注册配置文件适配器
func RegisterClient(resolver ClientResover) {
	proto := resolver.Name()
	if _, ok := clientResolvers[proto]; ok {
		panic(fmt.Errorf("xrpc: 不能重复注册:%s", proto))
	}
	clientResolvers[proto] = resolver
}

// Deregister 清理配置适配器
func DeregisterClient(name string) {
	delete(clientResolvers, name)
}

// newXRPC 根据适配器名称及参数返回配置处理器
func newClient(name string, setting config.Config) (Client, error) {
	val := setting.Value("proto")
	proto := val.String()
	resolver, ok := clientResolvers[proto]
	if !ok {
		return nil, fmt.Errorf("xrpc: 未知的协议类型:%s", proto)
	}
	return resolver.Resolve(name, setting)
}

func IsIpPortAddr(addr string) (ip string, port int, ok bool) {

	parties := strings.SplitN(addr, ":", 2)
	if len(parties) <= 1 {
		ok = false
		return
	}
	ok = net.ParseIP(parties[0]).To4() != nil
	if !ok {
		return
	}
	ip = parties[0]

	port, err := strconv.Atoi(parties[1])
	if err != nil {
		ok = false
		return
	}
	ok = true
	return
}
