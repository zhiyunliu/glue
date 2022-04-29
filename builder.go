package gel

import (
	"fmt"
	"sync"

	"github.com/zhiyunliu/gel/cache"
	"github.com/zhiyunliu/gel/container"
	"github.com/zhiyunliu/gel/dlocker"
	"github.com/zhiyunliu/gel/queue"
	"github.com/zhiyunliu/gel/xdb"
	"github.com/zhiyunliu/gel/xrpc"
)

var (
	contanier = container.NewContainer()

	lock sync.Mutex
)

var _ = registrydefault()
var builderMap = sync.Map{}

//注册标准对象的builder
func Registry(builder container.StandardBuilder) {
	builderMap.Store(builder.Name(), builder)
}

func getStandardInstance(name string) interface{} {
	instanceName := fmt.Sprintf("instance_%s", name)
	if obj, ok := builderMap.Load(instanceName); ok {
		return obj
	}
	lock.Lock()
	defer lock.Unlock()

	tmp, ok := builderMap.Load(name)
	if !ok {
		panic(fmt.Errorf("未注册builder.name=%s", name))
	}
	obj := tmp.(container.StandardBuilder).Build(contanier)

	builderMap.Store(instanceName, obj)
	return obj
}

//注册默认的提供程序
func registrydefault() error {
	Registry(xdb.NewBuilder())
	Registry(cache.NewBuilder())
	Registry(queue.NewBuilder())
	Registry(xrpc.NewBuilder())
	Registry(dlocker.NewBuilder())
	return nil
}
