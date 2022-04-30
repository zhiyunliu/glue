package server

import (
	"github.com/zhiyunliu/gel/middleware"
	"github.com/zhiyunliu/gel/middleware/auth/jwt"
	"github.com/zhiyunliu/gel/middleware/metrics"
	"github.com/zhiyunliu/gel/middleware/ratelimit"
	"github.com/zhiyunliu/gel/middleware/tracing"
)

func init() {
	middleware.Registry(jwt.NewBuilder())
	middleware.Registry(metrics.NewBuilder())
	middleware.Registry(ratelimit.NewBuilder())
	middleware.Registry(tracing.NewBuilder())
}
