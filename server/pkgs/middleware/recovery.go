package middleware

import (
	"fmt"
	"net/http"

	"github.com/zhiyunliu/velocity/context"

	"github.com/micro-plat/hydra/global"
)

//Recovery 用于处理请求过程中出现的非预见的错误
//服务器首个Recovery中间件，应使用 Recovery(true)才能正确释放context资源
func Recovery() context.Handler {
	return func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				path := ctx.Request().Path().GetURL().Path
				ctx.Log().Info(serverType+".recovery:", ctx.Request().Path().GetMethod(), path, "from", ctx.User().GetClientIP())
				ctx.Log().Errorf("-----[Recovery] panic recovered:/n%s/n%s", err, global.GetStack())
				ctx.Response().Abort(http.StatusNotExtended, fmt.Errorf("%v", "Server Error"))
			}
			if len(needPrt) > 0 && needPrt[0] {
				//5.释放资源
				ctx.Close()
			}

		}()
		ctx.Next()
	}
}
