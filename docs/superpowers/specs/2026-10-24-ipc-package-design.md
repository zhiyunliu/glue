# IPC 包设计

## 目标与范围

在当前仓库的 `ipc/` 目录实现可复用的跨进程通信包，遵循根目录 Go module，不创建嵌套模块。保留已有的 `ipc/xx.md`，增加包实现、自动化测试和 `ipc/example/` 下的 server/client 示例。

支持通过本地传输执行命令请求/响应，并提供命名共享内存与跨进程锁 API。平台实现覆盖 Linux、macOS 和 Windows。

## 结构与 API

- `ipc.go`：导出统一的监听器、共享内存、进程锁 API。
- `contract.go`：请求/响应、状态码与 Handler 契约。
- `protocol.go`：长度前缀 JSON 帧的编码、解码与帧大小限制。
- `server.go`、`client.go`：命令注册/分发、请求响应和连接生命周期。
- `transport_*.go`：Unix Domain Socket 和 Windows 命名管道。
- `shm_*.go`：POSIX 共享内存和 Windows 文件映射。
- `lock_*.go`：Unix 文件锁和 Windows 命名互斥锁。
- `ipc/example/`：演示命令调用及共享内存访问。

请求由命令名和 JSON data 组成。Handler 返回固定字段的响应；业务失败由响应状态表示，传输或协议失败通过 Go error 返回。单个客户端连接的请求响应按序执行。

## 可靠性与资源安全

- 帧读写必须完整处理短读/短写，并拒绝超过最大帧大小的数据。
- 关闭服务端时，必须解除阻塞中的 Accept 并关闭活动连接，等待连接处理结束。
- 共享内存创建不能意外截断或重置已有区域；输入大小和映射失败应显式返回错误。
- 共享内存及锁的释放由调用者负责；示例须检查关键的加锁、解锁和关闭错误。
- Handler panic 应转换为统一的失败响应，不影响其他连接处理。

## 验证

- 为协议边界、请求响应编解码、命令分发及关闭生命周期添加测试。
- 在 Windows 当前环境运行 `go test ./ipc/...`。
- 交叉编译检查 Linux、macOS、Windows 平台实现是否能编译。
- 示例须在仓库根模块内可编译，并保持原有包行为不变。
