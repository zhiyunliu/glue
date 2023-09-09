package engine

import "github.com/zhiyunliu/glue/contrib/alloter"

type AlloterEngine interface {
	HandleRequest(r alloter.IRequest, resp alloter.ResponseWriter) (err error)
}
