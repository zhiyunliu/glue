package container

import (
	"fmt"
	"strings"
	"sync"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/zhiyunliu/velocity/config"
	"github.com/zhiyunliu/velocity/global"
)

//ICloser 关闭
type ICloser interface {
	Close() error
}

type CreateFunc func(setting config.Config) (interface{}, error)

//IContainer 组件容器
type IContainer interface {
	GetOrCreate(typeName, name string, creator CreateFunc, keys ...string) (interface{}, error)
	Remove(typeName, name string, keys ...string) error
	ICloser
}

//Container 容器用于缓存公共组件
type Container struct {
	mutex sync.Mutex
	cache cmap.ConcurrentMap
}

//NewContainer 构建容器,用于管理公共组件
func NewContainer() *Container {
	c := &Container{
		cache: cmap.New(),
	}
	return c

}

//GetOrCreate 获取指定名称的组件，不存在时自动创建
func (c *Container) GetOrCreate(typeName string, name string, creator CreateFunc, keys ...string) (interface{}, error) {

	typeSetting := global.Setting.Get(typeName)
	if typeSetting == nil {
		return nil, fmt.Errorf("类型:%s 未进行配置", typeName)
	}

	nameSetting := typeSetting.Get(name)
	if typeSetting == nil {
		return nil, fmt.Errorf("类型:%s name=%s.未进行配置", typeName, name)
	}

	//2. 根据配置创建组件
	key := fmt.Sprintf("%s_%s_%s", typeName, name, strings.Join(keys, "_"))
	val, ok := c.cache.Get(key)
	if ok {
		return val, nil
	}
	if creator == nil {
		return nil, fmt.Errorf("未设置Type:%s,Name:%s，构建函数", typeName, name)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	val, err := creator(nameSetting)
	if err != nil {
		return nil, fmt.Errorf("创建Type:%s,Name:%s，失败，Error:%+v", typeName, name, err)
	}
	c.cache.Set(key, val)
	return val, nil
}

//Close 释放组件资源
func (c *Container) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for item := range c.cache.IterBuffered() {
		if closer, ok := item.Val.(ICloser); ok {
			closer.Close()
		}
	}
	c.cache.Clear()
	return nil
}

//Remove 释放组件资源
func (c *Container) Remove(typeName, name string, keys ...string) error {
	key := fmt.Sprintf("%s_%s_%s", typeName, name, strings.Join(keys, "_"))
	c.cache.RemoveCb(key, func(key string, v interface{}, exists bool) bool {
		if !exists {
			return false
		}
		if closer, ok := v.(ICloser); ok {
			closer.Close()
		}
		return true
	})
	return nil
}
