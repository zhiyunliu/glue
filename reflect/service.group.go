package reflect

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/zhiyunliu/velocity/middleware"
)

/*
1. ---------------------------------------
/xx/xx  func get/post ==>
group.path = /xx/xx
group.handing=nil
group.handed=nil
group.parent = nil
group.children = []
group.service[get]= {name:get,handing,handle,handled}
group.service[post]= {name:post,handing,handle,handled}
2. ----------------------------------------
/xx/xx  obj(handing,handed,handle,ahandle) get/post ==>
group.path = /xx/xx
group.handing=handing
group.handed=handed
group.parent = nil
group.children = [{
	group.path = a
	group.parent = group
	group.service[get]= {name:get,handle}
	group.service[post]= {name:post,handle}
}]
group.service[get]= {name:get,handing,handle,handled}
group.service[post]= {name:get,handing,handle,handled}
3. ----------------------------------------
/xx/xx  obj(ahanding,ahanded,ahandle) get/post ==>
group.path = /xx/xx
group.handing=nil
group.handed=nil
group.parent = nil
group.children = [{
	group.path = a
	group.handing=handing
	group.handed=handed
	group.parent = group
	group.service[get]= {name:get,handle}
	group.service[post]= {name:post,handle}
}]
group.service={}
*/

type ServiceGroup struct {
	Path     string //原始注册服务路径
	Handling middleware.Handler
	Handled  middleware.Handler
	Children map[string]*ServiceGroup
	Services map[string]*ServiceUnit
	parent   *ServiceGroup
	methods  []string
}

type ServiceUnit struct {
	Name     string
	Handling middleware.Handler
	Handled  middleware.Handler
	Handle   middleware.Handler
	Group    *ServiceGroup
}

func newServiceGroup(path string, methods ...string) *ServiceGroup {
	if len(methods) == 0 {
		methods = []string{http.MethodGet, http.MethodPost}
	}
	return &ServiceGroup{
		Path:     path,
		methods:  methods,
		Services: make(map[string]*ServiceUnit),
		Children: make(map[string]*ServiceGroup),
	}
}

func (g *ServiceGroup) GetChild(name string) *ServiceGroup {
	child, ok := g.Children[name]
	if ok {
		return child
	}

	child = &ServiceGroup{
		Path:     name,
		parent:   g,
		methods:  g.methods,
		Services: make(map[string]*ServiceUnit),
		Children: make(map[string]*ServiceGroup, 0),
	}
	g.Children[name] = child
	return child
}

func (g *ServiceGroup) AddHandle(subName string, handler middleware.Handler) {
	if strings.EqualFold(subName, "") {
		for _, m := range g.methods {
			g.Services[m] = &ServiceUnit{
				Name:   m,
				Handle: handler,
				Group:  g,
			}
		}
	} else {
		child := g.GetChild(subName)
		child.AddHandle("", handler)
	}
}

func (g *ServiceGroup) AddHandling(subName string, handler middleware.Handler) {
	if strings.EqualFold(subName, "") {
		for _, m := range g.methods {
			g.Services[m] = &ServiceUnit{
				Name:     m,
				Handling: handler,
				Group:    g,
			}
		}
	} else {
		child := g.GetChild(subName)
		child.AddHandle("", handler)
	}
}

func (g *ServiceGroup) AddHandled(subName string, handler middleware.Handler) {
	if strings.EqualFold(subName, "") {
		for _, m := range g.methods {
			g.Services[m] = &ServiceUnit{
				Name:    m,
				Handled: handler,
				Group:   g,
			}
		}
	} else {
		child := g.GetChild(subName)
		child.AddHandle("", handler)
	}
}

func (g *ServiceGroup) HasService() bool {
	return len(g.Services) > 0
}

func (g *ServiceGroup) HasChildren() bool {
	return len(g.Children) > 0
}

func (g *ServiceGroup) IsValid() error {
	if !(g.HasService() && g.HasChildren()) {
		return fmt.Errorf("%s无可用注册处理函数", g.Path)
	}
	errs := []error{}
	for m, s := range g.Services {
		if s.Handle == nil {
			errs = append(errs, fmt.Errorf("服务地址：%s ,Method:%s,没有提供处理函数", g.GetReallyPath(), m))
		}
	}

	for _, c := range g.Children {
		if err := c.IsValid(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		errStrs := make([]string, len(errs))
		for i := range errs {
			errStrs[i] = errs[i].Error()
		}
		return fmt.Errorf(strings.Join(errStrs, "\n"))
	}
	return nil
}
func (g *ServiceGroup) GetReallyPath() string {
	if g.parent != nil {
		return fmt.Sprintf("%s/%s", g.parent.GetReallyPath(), g.Path)
	}
	return g.Path
}
