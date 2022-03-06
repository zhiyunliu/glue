package middleware

import "github.com/gin-gonic/gin"

type GinContext gin.Context

func (ctx *GinContext) Trace(...interface{}) {

}

func (ctx *GinContext) Close() {

}
