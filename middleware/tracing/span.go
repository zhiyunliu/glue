package tracing

import (
	"fmt"
	"net"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/metadata"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

func setSpanAttrs(ctx context.Context, span trace.Span, m interface{}) {
	attrs := []attribute.KeyValue{}
	var remote string
	serverType := ctx.ServerType()

	method := ctx.Request().GetMethod()
	route := ctx.Request().Path().FullPath()
	path := ctx.Request().Path().GetURL().Path

	attrs = append(attrs, buildKey(serverType, "method").String(method))
	attrs = append(attrs, buildKey(serverType, "route").String(route))
	attrs = append(attrs, buildKey(serverType, "target").String(path))
	attrs = append(attrs, buildKey(serverType, "system").String(fmt.Sprintf("%s.%s", global.AppName, ctx.ServerName())))

	remote = ctx.Request().GetClientIP()
	attrs = append(attrs, buildKey(serverType, "client-ip").String(remote))
	//attrs = append(attrs, peerAttr(remote)...)
	if md, ok := metadata.FromServerContext(ctx.Context()); ok {
		attrs = append(attrs, semconv.PeerServiceKey.String(md.Get(serviceHeader)))
	}

	span.SetAttributes(attrs...)
}

// peerAttr returns attributes about the peer address.
func peerAttr(addr string) []attribute.KeyValue {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return []attribute.KeyValue(nil)
	}

	if host == "" {
		host = "127.0.0.1"
	}

	return []attribute.KeyValue{
		semconv.NetPeerIPKey.String(host),
		semconv.NetPeerPortKey.String(port),
	}
}

func buildKey(serverType, key string) attribute.Key {
	return attribute.Key(fmt.Sprintf("%s.%s", serverType, key))
}
