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

const (
	MethodPost   = "POST"
	MethodGet    = "GET"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
)

var methodMap = map[string]bool{
	MethodGet:    true,
	MethodPost:   true,
	MethodPut:    true,
	MethodDelete: true,
}

func adjustMethods(methods ...Method) []Method {
	if len(methods) == 0 {
		return []Method{MethodGet, MethodPost, MethodPut, MethodDelete}
	}
	result := []Method{}
	for _, v := range methods {
		if !isValidMethod(v) {
			continue
		}
		result = append(result, v)
	}
	return result
}

func isValidMethod(method Method) bool {
	_, ok := methodMap[strings.ToUpper(string(method))]
	return ok
}

type RouterGroup struct {
	basePath      string
	middlewares   []middleware.Middleware
	ServiceGroups map[string]*router.Group
	Children      map[string]*RouterGroup
}

func NewRouterGroup(basePath string) *RouterGroup {
	return &RouterGroup{
		basePath:      basePath,
		ServiceGroups: make(map[string]*router.Group),
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
		ServiceGroups: make(map[string]*router.Group),
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
func (group *RouterGroup) Handle(relativePath string, handler interface{}, methods ...Method) {
	methods = adjustMethods(methods...)

	mths := make([]string, len(methods))
	for i := range methods {
		mths[i] = string(methods[i])
	}

	svcGroup, err := router.ReflectHandle(group.basePath, relativePath, handler, mths...)
	if err != nil {
		log.Error(err)
		return
	}

	if _, ok := group.ServiceGroups[relativePath]; ok {
		absolutePath := group.calculateAbsolutePath(relativePath)
		log.Error(fmt.Errorf("存在相同路径注册:%s", absolutePath))
		return
	}
	group.ServiceGroups[relativePath] = svcGroup
}

func (group *RouterGroup) combineHandlers(middlewares ...middleware.Middleware) []middleware.Middleware {
	finalSize := len(group.middlewares) + len(middlewares)

	mergedHandlers := make([]middleware.Middleware, finalSize)
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
