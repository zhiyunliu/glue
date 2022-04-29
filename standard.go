package gel

import (
	"github.com/zhiyunliu/gel/cache"
	"github.com/zhiyunliu/gel/dlocker"
	"github.com/zhiyunliu/gel/queue"
	"github.com/zhiyunliu/gel/xdb"
	"github.com/zhiyunliu/gel/xrpc"
)

//DB 获取DB 处理对象
func DB() xdb.StandardDB {
	obj := getStandardInstance(xdb.DbTypeNode)
	return obj.(xdb.StandardDB)
}

//Cache 获取Cache 处理对象
func Cache() cache.StandardCache {
	obj := getStandardInstance(cache.CacheTypeNode)
	return obj.(cache.StandardCache)
}

//Queue 获取Queue 处理对象
func Queue() queue.StandardQueue {
	obj := getStandardInstance(queue.QueueTypeNode)
	return obj.(queue.StandardQueue)
}

//RPC 获取RPC 处理对象
func RPC() xrpc.StandardRPC {
	obj := getStandardInstance(xrpc.TypeNode)
	return obj.(xrpc.StandardRPC)
}

//DLocker 获取DLocker 处理对象
func DLocker() dlocker.DLocker {
	obj := getStandardInstance(dlocker.TypeNode)
	return obj.(dlocker.DLocker)
}

//暂时没考虑用泛型
func Custom(name string) interface{} {
	obj := getStandardInstance(name)
	return obj
}
