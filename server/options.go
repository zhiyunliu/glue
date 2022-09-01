package server

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
)

type PortPicker func(begin, end, step int64) int64

var (
	rnd      = rand.New(rand.NewSource(time.Now().Unix()))
	portPick = map[string]PortPicker{
		"rand": randPortPicker,
		"seq":  seqPortPicker,
	}
)

type Option func(*options)

type options struct {
	SrvType         string
	SrvName         string
	RequestDecoder  DecodeRequestFunc  //:          server.DefaultRequestDecoder,
	ResponseEncoder EncodeResponseFunc //:          server.DefaultResponseEncoder,
	ErrorEncoder    EncodeErrorFunc    //:          server.DefaultErrorEncoder,
}

func setDefaultOptions() *options {
	return &options{
		SrvType:         "api",
		RequestDecoder:  DefaultRequestDecoder,
		ResponseEncoder: DefaultResponseEncoder,
		ErrorEncoder:    DefaultErrorEncoder,
	}
}

func WithSrvType(srvType string) Option {
	return func(o *options) {
		o.SrvType = srvType
	}
}

func WithSrvName(name string) Option {
	return func(o *options) {
		o.SrvName = name
	}
}
func WithRequestDecoder(requestDecoder DecodeRequestFunc) Option {
	return func(o *options) {
		o.RequestDecoder = requestDecoder
	}
}

func WithResponseEncoder(responseEncoder EncodeResponseFunc) Option {
	return func(o *options) {
		o.ResponseEncoder = responseEncoder
	}
}

func WithErrorEncoder(errorEncoder EncodeErrorFunc) Option {
	return func(o *options) {
		o.ErrorEncoder = errorEncoder
	}
}

//:8080
//2000-30000
//2000-30000:rand
//2000-30000:seq:2
func GetAvaliableAddr(addr string) (newAddr string) {
	//没有指定范围
	if !strings.Contains(addr, "-") {
		return addr
	}
	method := "rand"
	step := int64(1)
	parties := strings.SplitN(addr, ":", 3)
	if len(parties) == 2 {
		addr = parties[0]
		method = parties[1]
	}
	if len(parties) == 3 {
		addr = parties[0]
		method = parties[1]
		step, _ = strconv.ParseInt(parties[2], 10, 32)
	}

	parties = strings.SplitN(addr, "-", 2)
	begin, err := strconv.ParseInt(parties[0], 10, 32)
	if err != nil || begin <= 0 {
		panic(fmt.Errorf("指定端口配置错误:%s", addr))
	}
	end, err := strconv.ParseInt(parties[1], 10, 32)
	if err != nil || end <= 0 || end < begin {
		panic(fmt.Errorf("指定端口配置错误:%s", addr))
	}
	call, ok := portPick[method]
	if !ok {
		panic(fmt.Errorf("指定端口配置错误:%s", addr))
	}
	np := call(begin, end, step)
	if np == 0 {
		panic(fmt.Errorf("未获取到有效的端口,请检查配置:%s", addr))
	}
	newAddr = fmt.Sprintf(":%d", np)
	return
}

func randPortPicker(begin, end, step int64) int64 {
	for {
		np := rnd.Int63n(end-begin) + begin
		log.Debugf("检测端口(rand):%d", np)
		if !ScanPort("TCP", global.LocalIp, np) {
			return np
		}
	}
}

func seqPortPicker(begin, end, step int64) int64 {
	for np := begin; np < end; np++ {
		log.Debugf("检测端口(seq):%d", np)
		if !ScanPort("TCP", global.LocalIp, np) {
			return np
		}
	}
	return 0
}

func ScanPort(protocol string, hostname string, port int64) bool {
	p := strconv.FormatInt(port, 10)
	addr := net.JoinHostPort(hostname, p)
	conn, err := net.DialTimeout(protocol, addr, 1*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
