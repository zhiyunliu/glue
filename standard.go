package glue

import (
	"github.com/zhiyunliu/glue/cache"
	"github.com/zhiyunliu/glue/circuitbreaker"
	"github.com/zhiyunliu/glue/dlocker"
	"github.com/zhiyunliu/glue/metrics"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/glue/ratelimit"
	"github.com/zhiyunliu/glue/standard"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/glue/xhttp"
	"github.com/zhiyunliu/glue/xrpc"
)

// DB 获取DB 处理对象
func DB(name string, opts ...xdb.Option) xdb.IDB {
	obj := standard.GetInstance(xdb.DbTypeNode)
	return obj.(xdb.StandardDB).GetDB(name, opts...)
}

// Cache 获取Cache 处理对象
func Cache(name ...string) cache.ICache {
	obj := standard.GetInstance(cache.TypeNode)
	return obj.(cache.StandardCache).GetCache(name...)
}

// Queue 获取Queue 处理对象
func Queue(name ...string) queue.IQueue {
	obj := standard.GetInstance(queue.TypeNode)
	return obj.(queue.StandardQueue).GetQueue(name...)
}

// RPC 获取RPC 处理对象
func RPC(name ...string) xrpc.Client {
	obj := standard.GetInstance(xrpc.TypeNode)
	return obj.(xrpc.StandardRPC).GetRPC(name...)
}

// Http 获取Http处理对象
func Http(name ...string) xhttp.Client {
	obj := standard.GetInstance(xhttp.TypeNode)
	return obj.(xhttp.StandardHttp).GetHttp(name...)
}

// DLocker 获取DLocker 处理对象
func DLocker(key string) dlocker.DLocker {
	obj := standard.GetInstance(dlocker.TypeNode)
	return obj.(dlocker.StandardLocker).GetDLocker().Build(key)
}

// 暂时没考虑用泛型
func Custom(name string) interface{} {
	obj := standard.GetInstance(name)
	return obj
}

// 注册默认的提供程序
func init() {
	standard.Registry(xdb.NewBuilder())
	standard.Registry(cache.NewBuilder())
	standard.Registry(queue.NewBuilder())
	standard.Registry(xrpc.NewBuilder())
	standard.Registry(xhttp.NewBuilder())
	standard.Registry(dlocker.NewBuilder())
	standard.Registry(metrics.NewBuilder())
	standard.Registry(ratelimit.NewBuilder())
	standard.Registry(circuitbreaker.NewBuilder())
}
