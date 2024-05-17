package engine

import (
	"fmt"
	"path"
	"strings"

	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/router"
)

type Method string

func (m Method) String() string {
	return string(m)
}

func (m Method) Apply(opts *RouterOptions) {
	opts.Methods = append(opts.Methods, string(m))
}

const (
	MethodPost   Method = "POST"
	MethodGet    Method = "GET"
	MethodPut    Method = "PUT"
	MethodDelete Method = "DELETE"
)

var methodMap = map[Method]Method{
	MethodGet:    MethodGet,
	MethodPost:   MethodPost,
	MethodPut:    MethodPut,
	MethodDelete: MethodDelete,
}

func adjustMethods(methods ...string) []string {
	if len(methods) == 0 {
		return []string{string(MethodGet), string(MethodPost), string(MethodPut), string(MethodDelete)}
	}
	resultMap := map[Method]struct{}{}
	for _, v := range methods {
		if mth, ok := isValidMethod(v); ok {
			resultMap[mth] = struct{}{}
		}
	}
	result := make([]string, len(resultMap))
	i := 0
	for k := range resultMap {
		result[i] = string(k)
		i++
	}
	return methods
}

func isValidMethod(orgMethod string) (method Method, ok bool) {
	method, ok = methodMap[Method(strings.ToUpper(orgMethod))]
	return
}

type RouterWrapper struct {
	*router.Group
	opts *RouterOptions
}

type RouterGroup struct {
	basePath      string
	middlewares   []middleware.Middleware
	ServiceGroups map[string]*RouterWrapper
	Children      map[string]*RouterGroup
}

func NewRouterGroup(basePath string) *RouterGroup {
	return &RouterGroup{
		basePath:      basePath,
		ServiceGroups: make(map[string]*RouterWrapper),
		Children:      make(map[string]*RouterGroup),
	}
}

// Use adds middleware to the group, see example code in GitHub.
func (group *RouterGroup) Use(middlewares ...middleware.Middleware) *RouterGroup {
	newmids := make([]middleware.Middleware, 0)
	for i := range middlewares {
		if middlewares[i] != nil {
			newmids = append(newmids, middlewares[i])
		}
	}

	group.middlewares = append(group.middlewares, newmids...)
	return group
}

// Group creates a new router group. You should add all the routes that have common middlewares or the same path prefix.
// For example, all the routes that use a common middleware for authorization could be grouped.
func (group *RouterGroup) Group(relativePath string, middlewares ...middleware.Middleware) *RouterGroup {
	child := &RouterGroup{
		middlewares:   group.combineHandlers(middlewares...),
		basePath:      group.calculateAbsolutePath(relativePath),
		ServiceGroups: make(map[string]*RouterWrapper),
		Children:      make(map[string]*RouterGroup),
	}
	group.Children[relativePath] = child
	return child
}

// BasePath returns the base path of router group.
// For example, if v := router.Group("/rest/n/v1/api"), v.BasePath() is "/rest/n/v1/api".
func (group *RouterGroup) BasePath() string {
	return group.basePath
}

// Handle registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
// See the example code in GitHub.
//
// For GET, POST requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (group *RouterGroup) Handle(relativePath string, handler interface{}, opts ...RouterOption) {
	ropts := &RouterOptions{}
	for _, opt := range opts {
		opt.Apply(ropts)
	}
	methods := ropts.Methods

	methods = adjustMethods(methods...)
	svcGroup, err := router.ReflectHandle(group.basePath, relativePath, handler, methods...)
	if err != nil {
		log.Error(err)
		return
	}

	if _, ok := group.ServiceGroups[relativePath]; ok {
		absolutePath := group.calculateAbsolutePath(relativePath)
		log.Error(fmt.Errorf("存在相同路径注册:%s", absolutePath))
		return
	}
	group.ServiceGroups[relativePath] = &RouterWrapper{
		Group: svcGroup,
		opts:  ropts,
	}
}

func (group *RouterGroup) combineHandlers(middlewares ...middleware.Middleware) []middleware.Middleware {
	mergedHandlers := make([]middleware.Middleware, len(group.middlewares)+len(middlewares))
	copy(mergedHandlers, group.middlewares)
	copy(mergedHandlers[len(group.middlewares):], middlewares)
	return mergedHandlers
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	relativePath = strings.TrimPrefix(relativePath, "/")
	relativePath = strings.TrimSuffix(relativePath, "/")

	if relativePath == "" {
		return group.basePath
	}

	return path.Join(group.basePath, relativePath)
}
