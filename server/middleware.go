package server

import (
	"github.com/zhiyunliu/gel/middleware"
	"github.com/zhiyunliu/gel/middleware/auth/jwt"
	"github.com/zhiyunliu/gel/middleware/metrics"
)

func init() {
	middleware.Registry(jwt.NewBuilder())
	middleware.Registry(metrics.NewBuilder())
	middleware.Registry(jwt.NewBuilder())
	middleware.Registry(jwt.NewBuilder())
	middleware.Registry(jwt.NewBuilder())
	middleware.Registry(jwt.NewBuilder())
	middleware.Registry(jwt.NewBuilder())
}
