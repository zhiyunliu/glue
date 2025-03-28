package engine

import (
	"github.com/zhiyunliu/glue/middleware"
	"github.com/zhiyunliu/glue/middleware/auth/jwt"
	"github.com/zhiyunliu/glue/middleware/ratelimit"
	"github.com/zhiyunliu/glue/middleware/tracing"
)

func init() {
	middleware.Registry(jwt.NewBuilder())
	middleware.Registry(ratelimit.NewBuilder())
	middleware.Registry(tracing.NewBuilder())
}
