package gel

import (
	"github.com/zhiyunliu/glue/cache"
	"github.com/zhiyunliu/glue/dlocker"
	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/glue/xhttp"
	"github.com/zhiyunliu/glue/xrpc"
)

//DB 获取DB 处理对象
func DB() xdb.StandardDB {
	obj := getStandardInstance(xdb.DbTypeNode)
	return obj.(xdb.StandardDB)
}

//Cache 获取Cache 处理对象
func Cache() cache.StandardCache {
	obj := getStandardInstance(cache.TypeNode)
	return obj.(cache.StandardCache)
}

//Queue 获取Queue 处理对象
func Queue() queue.StandardQueue {
	obj := getStandardInstance(queue.TypeNode)
	return obj.(queue.StandardQueue)
}

//RPC 获取RPC 处理对象
func RPC() xrpc.StandardRPC {
	obj := getStandardInstance(xrpc.TypeNode)
	return obj.(xrpc.StandardRPC)
}

//Http 获取Http处理对象
func Http() xhttp.StandardHttp {
	obj := getStandardInstance(xhttp.TypeNode)
	return obj.(xhttp.StandardHttp)
}

//DLocker 获取DLocker 处理对象
func DLocker() dlocker.DLockerBuilder {
	obj := getStandardInstance(dlocker.TypeNode)
	return obj.(dlocker.DLockerBuilder)
}

//暂时没考虑用泛型
func Custom(name string) interface{} {
	obj := getStandardInstance(name)
	return obj
}
