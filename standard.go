package gel

import (
	"sync"

	"github.com/zhiyunliu/gel/cache"
	"github.com/zhiyunliu/gel/container"
	"github.com/zhiyunliu/gel/queue"
	"github.com/zhiyunliu/gel/xdb"
)

var (
	contanier  = container.NewContainer()
	standCache *cache.StandardCache
	standQueue *queue.StandardQueue
	standDB    *xdb.StandardDB
)

var (
	lock sync.Mutex
)

func DB() *xdb.StandardDB {
	if standDB != nil {
		return standDB
	}
	lock.Lock()
	defer lock.Unlock()
	if standDB != nil {
		return standDB
	}
	standDB = xdb.NewStandardDB(contanier)
	return standDB
}

func Cache() *cache.StandardCache {
	if standCache != nil {
		return standCache
	}
	lock.Lock()
	defer lock.Unlock()
	if standCache != nil {
		return standCache
	}
	standCache = cache.NewStandardCache(contanier)
	return standCache
}

func Queue() *queue.StandardQueue {
	if standQueue != nil {
		return standQueue
	}
	lock.Lock()
	defer lock.Unlock()
	if standQueue != nil {
		return standQueue
	}
	standQueue = queue.NewStandardQueue(contanier)
	return standQueue
}

func Dlocker() {

}
