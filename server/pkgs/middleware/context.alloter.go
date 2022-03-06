package middleware

import (
	"github.com/zhiyunliu/velocity/server/pkgs/alloter"
)

type AlloterContext alloter.Context

func (ctx *AlloterContext) Trace(...interface{}) {

}

func (ctx *AlloterContext) Close() {

}
