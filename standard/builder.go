package standard

import (
	"fmt"
	"sync"

	"github.com/zhiyunliu/glue/cache"
	"github.com/zhiyunliu/glue/circuitbreaker"
	"github.com/zhiyunliu/glue/container"
	"github.com/zhiyunliu/glue/dlocker"
	"github.com/zhiyunliu/glue/metrics"
	"github.com/zhiyunliu/glue/ratelimit"

	"github.com/zhiyunliu/glue/queue"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/glue/xhttp"
	"github.com/zhiyunliu/glue/xrpc"
)

var (
	contanier = container.NewContainer()
	lock      sync.Mutex
)

var _ = registrydefault()
var builderMap = sync.Map{}

//注册标准对象的builder
func Registry(builder container.StandardBuilder) {
	builderMap.Store(builder.Name(), builder)
}

func GetInstance(name string) interface{} {
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
	Registry(xhttp.NewBuilder())
	Registry(dlocker.NewBuilder())
	Registry(metrics.NewBuilder())
	Registry(ratelimit.NewBuilder())
	Registry(circuitbreaker.NewBuilder())
	return nil
}
