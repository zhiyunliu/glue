# IPC 本地跨进程通信

`server/ipc` 包与 `server/api`、`server/cron` 等服务包并列，通过本地流式传输提供命令请求/响应，并提供命名共享内存和跨进程互斥锁。平台实现分别使用 Unix Domain Socket、Windows 命名管道、POSIX 共享内存、Windows 文件映射、`flock` 和 Windows 命名 Mutex。

## 文件结构

```text
server/ipc/
├── ipc.go
├── contract.go
├── protocol.go
├── config.go
├── options.go
├── engine.go
├── command_server.go
├── server.go
├── client.go
├── transport_unix.go
├── transport_windows.go
├── shm_unix.go
├── shm_linux.go
├── shm_darwin.go
├── shm_windows.go
├── lock_unix.go
├── lock_windows.go
└── example/
    ├── server/main.go
    └── client/main.go
```

## 通信约定

每帧由 4 字节大端序 payload 长度和 JSON payload 组成，单帧最大 64 MiB。请求格式为 `{"cmd":"命令名","data":{...}}`；响应字段为 `code`、`sub_code`、`msg` 和 `data`。

- `Client.Call` 的 Go error 表示传输、帧或 JSON 协议错误。
- 远端业务失败仍通过 `*Response` 返回；检查 `Response.IsOK()` 和状态码。
- 同一个 `Client` 会串行处理调用，以保证请求与响应顺序。
- 服务端会把 Handler panic 转换为 `CodeHandlerPanic` 响应。

## 框架接入

`New(name, opts...)` 仅创建服务，在 `Start` 时才监听。服务实现 `transport.Server` 和 `transport.Endpointer`，可直接传给 `glue.Server`：

```go
server := ipc.New("ipc", ipc.WithAddr(endpoint), ipc.WithServiceName("local-worker"))
server.Group("/v1").Handle("/echo", func(ctx context.Context) interface{} {
    var args struct { Value string `json:"value"` }
    if err := ctx.Bind(&args); err != nil {
        return err
    }
    return args
}, engine.MethodPost)
app := glue.NewApp(glue.Server(server))
if err := app.Start(); err != nil {
    return err
}
```

这里的 `context` 是 `github.com/zhiyunliu/glue/context`，`engine` 是框架同名包；其他导入为根包 `glue` 和 `server/ipc`。
框架配置路径与 API 一致，为 `servers.<name>`；配置会覆盖构造参数：

```json
{
  "servers": {
    "ipc": {
      "config": {
        "addr": "glue-ipc-example",
        "engine": "gin",
        "status": "start",
        "stop_maximum_timeout": 15
      },
      "middlewares": [],
      "header": {}
    }
  }
}
```

Linux/macOS 的 `addr` 应使用本地 socket 路径。`status: stop` 禁止启动。
默认使用真实 Gin 适配引擎和 `engine.RegistryEngineRoute`，不额外监听 HTTP 端口。
`Use`、`Group`、`Handle` 支持统一上下文、中间件、对象反射路由及默认追踪和恢复链。
命令按 POST 请求映射到路径，例如 `client.Call("/v1/echo", args)`；建议显式使用以 `/` 开头的命令路径。
框架返回值封装为 `ipc.OK(value)`；返回 `*ipc.Response` 则原样保留；错误转换为失败响应。
`RegisterCommand` 和 `RegisterFunc` 保持低层协议，优先精确匹配，不经过统一中间件链。

`Endpoint()` 返回注册用 `*url.URL`，其 `Path` 保存本地原始地址；客户端使用 `Address()` 获取原始地址，不应把 URL 字符串传给 `Dial`。
`transport.Transporter.Endpoint() string` 与该接口签名冲突，因此服务与 API 一样不实现 `Transporter`；当前无需独立适配对象，也不修改全局接口。

## 独立生命周期及兼容接口

```go
server := ipc.New("ipc", ipc.WithAddr(endpoint))
server.RegisterFunc("Get", func(req ipc.Request) *ipc.Response {
    return ipc.OK(map[string]int{"value": 42})
})
if err := server.Start(ctx); err != nil {
    return err
}
defer server.Close()
client, err := ipc.Dial(server.Address())
if err != nil {
    return err
}
defer client.Close()
response, err := client.Call("Get", nil)
```

`ctx` 是标准库上下文。`Start` 绑定成功后返回，重复启动返回 `ErrServerAlreadyServing`。
`Stop` 关闭监听器及活动连接，取消框架请求上下文，再等待处理器退出，等待受调用上下文与最大超时共同限制。
无法强制终止任意业务函数；超时返回后清理继续，完成之前禁止重新启动。
未启动时可以安全停止；停止完成可再次 `Start`，命令和路由保持不变。
`Close` 是兼容终止入口，之后 `Start` 返回 `net.ErrClosed`，`Serve` 安全返回；低层 `NewServer(endpoint)` 仍立即监听，配合阻塞的 `Serve()` 使用。
监听错误由 `Start` 返回，接收循环错误由 `Serve` 或后续 `Stop` 返回；启动回调失败会清理监听器和连接。

## 示例运行

在仓库根目录打开两个终端，分别运行：

```powershell
# 终端一
go run ./server/ipc/example/server
# 终端二
go run ./server/ipc/example/client
```

服务端示例注册 `UpdateValue`、`Reset` 和 `Get` 命令；客户端演示强类型绑定响应和未知命令处理。

## 共享内存和锁

`OpenShared(name, size, create)` 返回一段由调用方管理的共享字节区域。`create=true` 在区域不存在时创建；若名称已存在，则打开已有区域且要求其大小不小于 `size`，不会截断已有数据。调用 `SharedRegion.Close()` 会解除当前进程的映射，但不会删除命名共享内存对象。

- Linux 共享区域位于 POSIX 共享内存命名空间（通常由 `/dev/shm` 承载）；macOS 使用 POSIX `shm_open`。
- POSIX 共享内存对象会在所有映射关闭后继续存在。使用完成后需由应用显式清理；Linux 可删除对应 `/dev/shm/<name>`，macOS 可使用 `shm_unlink`。
- Windows 映射对象在最后一个句柄关闭后由系统回收。
- 共享内存只放固定布局数据，不应存储 Go 指针、`string`、slice、map 或 channel。
- 共享区域跨进程读写的同步由应用负责。示例通过 `ProcessLock` 串行化访问。
- Windows Mutex 由获得它的 OS 线程持有；同一次 `Lock`/`Unlock` 必须由同一个 goroutine 顺序调用。服务端示例在同一个 Handler 调用中完成加锁和解锁。
- Unix 锁使用 `flock` 文件；锁文件本身不会在 `Close` 时删除。
- Unix socket listener 启动时会在端点同目录创建一个隐藏的 `.ipc-listener-<hash>.lock` 文件，用于串行化残留 socket 清理和绑定；该文件会保留，不应在服务可能启动时手动删除。

## 平台注意事项

| 能力 | Linux | macOS | Windows |
| --- | --- | --- | --- |
| 命令通道 | Unix Domain Socket | Unix Domain Socket | 命名管道 |
| 共享内存 | POSIX 命名共享内存 | POSIX `shm_open` | 文件映射 |
| 跨进程锁 | `flock` | `flock` | 命名 Mutex |
| deadline | 由 socket 支持 | 由 socket 支持 | 命名管道不支持，调用返回 `ErrDeadlineUnsupported` |

Windows 端点参数只传管道名，不包含 `\\.\pipe\` 前缀。跨用户共享对象需要明确配置 Windows 对象命名空间及安全描述符；本包默认使用当前会话命名空间。

## 验证

```powershell
go test ./server/ipc/... -count=1
go test -race -vet=off ./server/ipc/... -count=1
go vet ./server/ipc/...
go vet -unsafeptr=false ./server/ipc/...
```

Windows 的完整 vet 检查仍报告既有 `shm_windows.go:82` 的 `unsafe.Pointer` 告警；关闭该单项检查后可验证其余静态检查。
