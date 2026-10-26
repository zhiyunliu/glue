package global

import "encoding/json"

type RouterList interface {
	GetType() string
	GetPathResult() any
}

var (
	//服务路由
	ServerRouterPathResult *RouterPathResult = NewRouterPathResult()
)

type RouterPathResult struct {
	serverRouterMap map[string]RouterList
}

func NewRouterPathResult() *RouterPathResult {
	return &RouterPathResult{
		serverRouterMap: map[string]RouterList{},
	}
}

// 存储服务的路由信息
func (r *RouterPathResult) Store(servername string, routerList RouterList) {
	r.serverRouterMap[servername] = routerList
}

// 迭代获取服务的路由信息
func (r *RouterPathResult) Range(callback func(k string, v RouterList) bool) {
	for k, v := range r.serverRouterMap {
		if !callback(k, v) {
			break
		}
	}
}

func (r *RouterPathResult) MarshalJSON() (bytes []byte, err error) {
	return json.Marshal(r.serverRouterMap)
}
