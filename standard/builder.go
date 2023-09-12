package standard

import (
	"fmt"
	"sync"

	"github.com/zhiyunliu/glue/container"
)

var (
	contanier  = container.NewContainer()
	lock       sync.Mutex
	builderMap = sync.Map{}
)

// 注册标准对象的builder
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

	//二次检查
	if obj, ok := builderMap.Load(instanceName); ok {
		return obj
	}

	tmp, ok := builderMap.Load(name)
	if !ok {
		panic(fmt.Errorf("未注册builder.name=%s", name))
	}
	obj := tmp.(container.StandardBuilder).Build(contanier)

	builderMap.Store(instanceName, obj)
	return obj
}
