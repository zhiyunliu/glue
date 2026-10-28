package xipc

import (
	sctx "context"

	"github.com/zhiyunliu/glue/config"
)

// Client 定义同步命令客户端。
type Client interface {
	//Request IPC请求，可通过context撤销请求
	//<param name="service">服务名称,ipc://srvName/command</param>
	//<param name="input">请求参数</param>
	//<param name="opts">请求选项</param>
	//<returns>响应体</returns>
	//<exception cref="error">请求失败</exception>
	//<param name="opts">请求选项</param>
	Request(ctx sctx.Context, path string, data any, opts ...RequestOption) (res Body, err error)

	//Call IPC调用，可通过context撤销调用
	Call(ctx sctx.Context, srvName string, command string, data any, opts ...RequestOption) (res Body, err error)
}

// ClientResover 定义配置文件转换方法
type ClientResolver interface {
	Name() string
	Resolve(proto string, cfg config.Config) (Client, error)
}
