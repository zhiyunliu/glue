package reflect

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/gel/middleware"
)

const Handling = "Handling"
const Handler = "Handle"
const Handled = "Handled"

type reflectCallback func(*ServiceGroup, string, middleware.Handler)

var funcCallback = map[string]reflectCallback{
	Handling: func(g *ServiceGroup, subName string, handler middleware.Handler) {
		g.AddHandling(subName, handler)
	},
	Handler: func(g *ServiceGroup, subName string, handler middleware.Handler) {
		g.AddHandle(subName, handler)
	},
	Handled: func(g *ServiceGroup, subName string, handler middleware.Handler) {
		g.AddHandled(subName, handler)
	},
}

var suffixList = []string{Handling, Handler, Handled}

//反射获取对象的Handle 方法
func ReflectHandle(path string, obj interface{}, method ...string) (*ServiceGroup, error) {
	//检查参数
	if path == "" || obj == nil {
		return nil, fmt.Errorf("注册对象路径和对象不能为空.Path:%s,Obj:%+v", path, obj)
	}

	//输入参数为函数
	group := newServiceGroup(path, method...)
	if sfunc, ok := getIsValidFunc(obj); ok {
		group.AddHandle("", sfunc)
		return group, nil
	}

	otype := reflect.TypeOf(obj)
	refval := reflect.ValueOf(obj)

	//检查对象类型
	if refval.Kind() != reflect.Ptr && refval.Kind() != reflect.Struct {
		return nil, fmt.Errorf("只能接收引用类型或struct; 实际是 %s", refval.Kind().String())
	}

	//reflect所有函数，检查函数签名
	cnt := otype.NumMethod()
	for i := 0; i < cnt; i++ {

		//检查函数参数是否符合接口要求
		funcName := otype.Method(i).Name
		method := refval.MethodByName(funcName)

		hasSuffix := checkFuncSuffix(funcName)
		if !hasSuffix {
			continue
		}

		//转换函数签名
		nf, ok := getIsValidFunc(method.Interface())
		if !ok {
			return nil, fmt.Errorf("函数【%s】是钩子类型(%v),但签名不是func(context.Context) interface{}", funcName, suffixList)
		}
		for _, sfx := range suffixList {
			if strings.HasSuffix(funcName, sfx) {
				subName := strings.ToLower(funcName[0 : len(funcName)-len(sfx)])
				funcCallback[sfx](group, subName, nf)
				break
			}
		}
	}
	if err := group.IsValid(); err != nil {
		return nil, err
	}

	return group, nil
}

func checkFuncSuffix(funcName string) bool {
	for i := range suffixList {
		if strings.HasSuffix(funcName, suffixList[i]) {
			return true
		}
	}
	return false
}

func getIsValidFunc(obj interface{}) (middleware.Handler, bool) {
	nfx, ok := obj.(func(context.Context) interface{})
	if ok {
		return middleware.Handler(nfx), true
	}
	return nil, false
}
