# IPC 客户端与共享能力

根 `ipc` 不依赖 `server/ipc`，仅负责客户端/服务端契约、协议注册工厂、配置、standard 和公共响应。
具体本地传输、framing、共享内存与进程锁位于 `contrib/ipc/default`（包名 `defaultipc`）。
服务端使用独立的 `server/ipc`；完整服务配置及平台限制见 [服务端说明](../server/ipc/README.md)。

## 标准入口

应用装载 `global.Config` 后，按名称读取 `ipcs` 节点：

```json
{
  "ipcs": {
    "default": {"addr": "glue-ipc-example"},
    "worker": {"addr": "glue-ipc-worker"}
  }
}
```

Windows 地址不包含命名管道前缀；Linux/macOS 使用本地 socket 路径。

```go
client := glue.IPC("")
response, err := client.Call("Get", nil)
if err != nil {
    return err
}
if !response.IsOK() {
    return fmt.Errorf("调用失败：%s", response.Msg)
}
var value struct { Value int64 `json:"value"` }
return response.Bind(&value)
```

导入根包 `github.com/zhiyunliu/glue`。`glue.IPC("worker", ipc.WithAddr(endpoint))` 可覆盖地址；
`ipc.WithOption(key, value)` 可覆盖配置项。选项中的值须能编码为 JSON；不支持的值会触发 panic。
配置错误或连接失败同其他标准组件一样触发 panic。
相同名称与相同最终选项复用客户端，选项顺序不影响标识，后设置的同名选项覆盖先前值。
空名称与 `default` 等价；不同名称或不同选项使用不同缓存对象。
客户端通过容器统一关闭；手动关闭会使所有持有该缓存实例的调用方无法继续调用，不会自动重连。

## 独立客户端

```go
client, err := ipc.NewClient("default", ipc.Config{Addr: endpoint})
if err != nil {
    return err
}
defer client.Close()
response, err := client.Call("Get", nil)
```

独立客户端由调用方负责关闭。同一连接上的请求串行执行，关闭可重复调用。
`Call` 的错误表示本地编码或传输失败；远端业务错误通过 `Response` 表达。
协议或传输错误会关闭损坏连接；关闭后返回 `ipc.ErrClosed`。
旧的客户端连接函数已删除，应迁移为 `glue.IPC` 或 `ipc.NewClient`。

## 共享能力

`defaultipc.OpenShared`、`defaultipc.OpenLock` 提供具体实现；根包仅保留 `SharedRegion`、`ProcessLock` 接口。
`defaultipc.Listen`/`defaultipc.Connect` 和 `defaultipc.ReadFrame`/`defaultipc.WriteFrame` 为底层传输与帧扩展提供公共能力。
这些低层 API 不经过标准容器缓存，资源由调用方管理。


## 验证与平台限制

Windows 可实际执行 `go test . ./ipc/... ./server/ipc/... -count=1`。
默认 `go vet` 会报告 Windows 映射地址转为字节切片时的 `unsafe.Pointer` 诊断；
这是底层映射 API 返回 `uintptr` 的边界，`go:nocheckptr` 仅控制运行时检查，不屏蔽 vet。
可运行 `go vet -unsafeptr=false ./ipc/... ./server/ipc/...` 检查其他分析项。
Linux/macOS 交叉编译不能替代对应系统的运行测试。

## 协议注册与破坏性迁移

直接使用根工厂或 standard 的程序必须导入协议实现：

```go
import (
    "github.com/zhiyunliu/glue/ipc"
    _ "github.com/zhiyunliu/glue/contrib/ipc/default"
)
```

`server/ipc` 内置空导入 default；根 `ipc` 不会反向导入任何实现。
旧 `ipc.NewClient(endpoint)` 已替换为 `ipc.NewClient(proto, Config, ...Option)`。
仅需具体默认客户端时也可调用 `defaultipc.NewClient(endpoint)`。
配置节点增加 `proto`，standard 未指定时使用 `default`；未知协议返回错误。
外部实现通过 `ipc.Register(ClientResolver)`、`ipc.RegisterServer(ServerResolver)` 注册。
服务端实现 `GetProto/GetAddr/SetHandler/Serve/Stop`；框架在 Serve 前安装传输无关 Handler。
注册应在应用初始化阶段完成；运行期间不要并发修改注册表。
