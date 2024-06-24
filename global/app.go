package global

type RouterList interface {
	GetType() string
	GetPathList() []string
}

var (
	//服务路由
	ServerRouterPathList *RouterPathList = NewRouterPathList()
)

type RouterPathList struct {
	serverRouterMap map[string][]RouterList
}

func NewRouterPathList() *RouterPathList {
	return &RouterPathList{
		serverRouterMap: map[string][]RouterList{},
	}
}

// 存储服务的路由信息
func (r *RouterPathList) Store(servername string, routerList RouterList) {
	r.serverRouterMap[servername] = append(r.serverRouterMap[servername], routerList)
}

// 迭代获取服务的路由信息
func (r *RouterPathList) Range(callback func(k string, v []RouterList) bool) {
	for k, v := range r.serverRouterMap {

		tmpv := make([]RouterList, len(v))
		copy(tmpv, v)
		if !callback(k, tmpv) {
			break
		}
	}
}
