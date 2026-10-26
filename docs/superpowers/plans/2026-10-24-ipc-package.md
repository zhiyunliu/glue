# IPC Package Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有仓库 `ipc/` 中交付一个可复用、跨平台的本地 IPC 包，包含命令调用、共享内存、进程锁和可运行示例。

**Architecture:** 包使用长度前缀 JSON 帧承载命令请求和响应；Unix 使用本地 socket、POSIX 共享内存及文件锁，Windows 使用命名管道、文件映射及命名互斥锁。公共 API 与平台实现分离；错误边界区分业务响应和传输错误。

**Tech Stack:** Go 1.25.0（仓库根模块），标准库 `net`、`encoding/json`、`sync`、`io`；复用仓库已有的 `golang.org/x/sys`。

## Global Constraints

- 使用仓库根模块 `github.com/zhiyunliu/glue`，不增加 `ipc/go.mod`。
- 保留现有 `ipc/xx.md`，不覆盖或移除用户提供的文档。
- 平台范围为 Linux、macOS、Windows。
- 新增 Go 代码中的注释使用简体中文。
- 调用系统 API、映射内存、关闭连接及锁操作的错误不得被静默丢弃。
- 共享内存创建不得截断已有区域；请求帧读写必须完整且受最大帧长度限制。
- Windows `Listener.Close` 必须解除阻塞中的 `Accept`；服务关闭须关闭活动连接并等待处理协程退出。

---

### Task 1: 建立 IPC 契约和帧协议

**Files:**
- Create: `ipc/ipc.go`
- Create: `ipc/contract.go`
- Create: `ipc/protocol.go`
- Create: `ipc/protocol_test.go`

**Interfaces:**
- Produces:
  - `Listen(endpoint string) (net.Listener, error)`
  - `SharedRegion`：`Bytes() []byte`、`Close() error`
  - `OpenShared(name string, size int, create bool) (SharedRegion, error)`
  - `ProcessLock`：`Lock() error`、`Unlock() error`、`Close() error`
  - `OpenLock(name string) (ProcessLock, error)`
  - `Request`：`Command() string`、`RawData() json.RawMessage`、`Bind(v any) error`
  - `Response`：`Code`、`SubCode`、`Msg`、`Data` 字段，JSON 标签分别为 `code`、`sub_code`、`msg`、`data`；以及 `IsOK()`、`Bind(v any) error`
  - `Handler`：`Handle(req Request) *Response`；`HandlerFunc` 实现该接口
  - 状态码常量：`CodeOK`、`CodeBadRequest`、`CodeUnknownCommand`、`CodeHandlerError`、`CodeHandlerPanic`、`CodeEncodeError`
  - `writeFrame(io.Writer, []byte) error`、`readFrame(io.Reader) ([]byte, error)`，帧头为 4 字节大端长度，最大帧 64 MiB。

- [ ] **Step 1: 编写协议测试**

  在 `protocol_test.go` 覆盖小帧往返、空 payload、短读、短写、超大 payload 和超大帧头。使用仅每次读写少量字节的 `io.Reader`/`io.Writer` 测试替身，确认 helper 会继续读写到完成，而不是假设一次 `Read` 或 `Write` 处理全部数据。

- [ ] **Step 2: 运行协议测试并确认失败**

  Run: `go test ./ipc -run 'Test(ReadFrame|WriteFrame)' -count=1`
  Expected: FAIL，目标函数尚未实现。

- [ ] **Step 3: 实现公共契约和帧编解码**

  使用 `io.ReadFull` 读帧头和 payload；写入时循环处理短写并将 `n <= 0` 且无错误的写入作为 `io.ErrShortWrite`。在分配 payload 前校验 64 MiB 上限。`Response.UnmarshalJSON` 保留原始 `data` 以供 `Bind` 做强类型解析；无 data 时 `Bind` 保持目标值不变。JSON 解析和编码错误原样返回。

- [ ] **Step 4: 运行协议测试**

  Run: `go test ./ipc -run 'Test(ReadFrame|WriteFrame)' -count=1`
  Expected: PASS；确认边界测试覆盖完整帧语义与大小限制。

### Task 2: 实现命令服务端与客户端

**Files:**
- Create: `ipc/server.go`
- Create: `ipc/client.go`
- Create: `ipc/client_server_test.go`

**Interfaces:**
- Consumes: Task 1 的 `Request`、`Response`、`Handler`、`wireRequest`（定义为含 `Cmd string` 与 `Data json.RawMessage` 的内部结构）及帧 helper。
- Produces:
  - `NewServer(endpoint string) (*Server, error)`
  - `(*Server).RegisterCommand(name string, h Handler)`
  - `(*Server).RegisterFunc(name string, f HandlerFunc)`
  - `(*Server).Serve() error`
  - `(*Server).Close() error`
  - `(*Server).Endpoint() string`
  - `Dial(endpoint string) (*Client, error)`
  - `(*Client).Call(command string, data any) (*Response, error)`
  - `(*Client).Close() error`
  - `ErrClosed`

- [ ] **Step 1: 编写端到端测试**

  使用平台监听器在临时端点启动 server goroutine。测试已注册命令成功绑定参数并返回 data、未知命令返回 `CodeUnknownCommand`、无参数命令、Handler panic 转换为 `CodeHandlerPanic`、无效请求 JSON 返回 `CodeBadRequest`、同一 client 连续调用，以及重复关闭 client/server。

- [ ] **Step 2: 运行端到端测试并确认失败**

  Run: `go test ./ipc -run 'Test(Server|Client)' -count=1`
  Expected: FAIL，`Server` 和 `Client` 尚未实现。

- [ ] **Step 3: 实现服务端分发和资源关闭**

  `Server` 使用读写锁保护 handler map，活动连接独立登记。关闭逻辑只执行一次：关闭 listener、关闭活动连接、等待连接处理退出；协调 `Serve` 的 Accept 和 `WaitGroup.Add`，避免 `Wait` 与新增任务并发。未知命令显式返回统一失败响应；Handler panic 转为 `CodeHandlerPanic`，不泄漏进程。响应 JSON 编码失败时返回编码错误响应，并检查第二次编码结果。

- [ ] **Step 4: 实现客户端调用**

  `Client.Call` 先编码请求，再以互斥锁串行执行帧写入及读取响应；业务非成功码仍返回 response 与 nil error，传输/协议/JSON 错误作为 Go error 返回。关闭后调用返回 `ErrClosed`。避免将被忽略的错误转换成成功响应。

- [ ] **Step 5: 运行端到端测试**

  Run: `go test ./ipc -run 'Test(Server|Client)' -count=1`
  Expected: PASS；关闭测试须证明 `Serve` 在关闭后退出，不遗留活动连接。

### Task 3: 实现 Unix socket 与 Windows 命名管道

**Files:**
- Create: `ipc/transport_unix.go`
- Create: `ipc/transport_windows.go`
- Create: `ipc/transport_test.go`

**Interfaces:**
- Consumes: Task 2 使用的私有函数 `listen(endpoint string) (net.Listener, error)` 与 `dialTransport(endpoint string) (net.Conn, error)`。
- Produces: Unix 平台的 Unix Domain Socket 监听/拨号；Windows 平台实现 `net.Listener` 和 `net.Conn` 的命名管道适配。

- [ ] **Step 1: 编写平台传输集成测试**

  覆盖 listener 接受连接、双向传输、关闭后 `Accept` 返回、重复关闭，以及服务端关闭时活动连接退出。Unix 使用 `t.TempDir()` 下的 socket 路径；Windows 使用带测试唯一后缀的管道名。

- [ ] **Step 2: 实现 Unix 传输**

  在 Unix 文件中以 `net.Listen("unix", endpoint)` 和 `net.Dial("unix", endpoint)` 实现。清理端点前用 `Lstat` 确认现有路径确为 socket；遇到普通文件、符号链接或清理失败时返回错误，不删除不相关文件。listener 关闭行为使用标准库实现。

- [ ] **Step 3: 实现 Windows 命名管道传输**

  在 Windows 文件中用 `golang.org/x/sys/windows` 创建双向 byte-mode 命名管道，适配 `net.Conn` 的读写和地址方法。listener 跟踪当前阻塞中的 pipe handle；`Close` 必须取消或关闭该 handle 以解除 `ConnectNamedPipe` 阻塞，随后令 `Accept` 返回关闭错误。处理 `ERROR_PIPE_CONNECTED` 为有效连接；其他 Win32 错误必须返回。由于系统 API 不提供 deadline，deadline 方法仅在文档明确说明不支持时返回可识别错误，不得伪装成功。

- [ ] **Step 4: 验证本机传输测试**

  Run: `go test ./ipc -run 'Test(Transport|Server)' -count=1`
  Expected: PASS；关闭中的 Accept 与活动连接均在限定时间内退出。

### Task 4: 实现共享内存与跨进程锁

**Files:**
- Create: `ipc/shm_unix.go`
- Create: `ipc/shm_linux.go`
- Create: `ipc/shm_darwin.go`
- Create: `ipc/shm_windows.go`
- Create: `ipc/lock_unix.go`
- Create: `ipc/lock_windows.go`
- Create: `ipc/shared_lock_unix_test.go`
- Create: `ipc/shared_lock_windows_test.go`
- Create: `ipc/shared_lock_test.go`
- Modify: `go.mod`

**Interfaces:**
- Consumes: Task 1 的 `SharedRegion`、`ProcessLock` 及统一入口。
- Produces: Linux `/dev/shm` 文件映射与 macOS `shm_open`/`mmap`、Unix `flock` 实现；Windows `CreateFileMapping`/`MapViewOfFile` 与命名 Mutex 实现。

- [ ] **Step 1: 编写共享内存与锁测试**

  测试非法名称/非正 size 返回错误，映射后字节可读写，`Close` 可重复调用且正确释放资源；测试锁成功获取/释放、竞争等待及关闭后错误。共享对象使用测试唯一名称；POSIX 测试在结束时移除命名对象。

- [ ] **Step 2: 实现 POSIX 共享内存和锁**

  `create=true` 时用 `O_CREAT|O_EXCL` 判断是否新建；仅新建对象允许设置大小。若对象已存在，检查文件长度至少为请求 size，不得 `Ftruncate` 覆盖既有内容。Linux 通过 `/dev/shm/<name>` 打开 POSIX 共享内存对象；macOS 用 `SYS_SHM_OPEN` 和 `SYS_SHM_UNLINK`。`create=false` 只打开已有对象并同样检查长度。所有 FD、mmap 和 flock 调用错误显式返回；映射关闭使用 `Munmap`。将 `golang.org/x/sys` 调整为 go.mod 直接依赖。

- [ ] **Step 3: 实现 Windows 文件映射和锁**

  创建/打开命名文件映射，检查句柄、视图地址和大小，确保失败时关闭已创建资源；Close 先解除映射再关闭句柄，并返回第一个资源释放错误。Mutex 的 `WaitForSingleObject` 检查等待结果和错误；`WAIT_ABANDONED` 表示已取得锁且应作为成功处理；`Unlock`、`Close` 错误需返回调用者。

- [ ] **Step 4: 运行本机共享资源测试**

  Run: `go test ./ipc -run 'Test(Shared|Lock)' -count=1`
  Expected: PASS；特别确认已有共享区域的内容与长度不会因再次 `OpenShared(..., true)` 被清零或截断。

### Task 5: 增加示例与使用说明

**Files:**
- Create: `ipc/example/server/main.go`
- Create: `ipc/example/client/main.go`
- Create: `ipc/README.md`

**Interfaces:**
- Consumes: 公共 API `NewServer`、`RegisterFunc`、`Dial`、`Call`、`OpenShared`、`OpenLock` 与 `Response.Bind`。
- Produces: 在仓库根 module 内可运行的 server/client 演示，演示 `UpdateValue`、`Reset`、`Get` 和未知命令。

- [ ] **Step 1: 编写示例使用说明**

  说明平台能力、控制面与数据面、端点命名、共享内存生命周期、锁的获取/释放、错误码、执行命令和 Windows deadline 限制。明确 `Close` 不等价于删除命名共享内存对象，并给出 POSIX 清理方式。

- [ ] **Step 2: 编写服务端和客户端示例**

  服务端注册三个示例命令；`UpdateValue` 在锁保护下写入共享内存，`Get` 读取值，`Reset` 演示无参命令。客户端验证传输 error 与业务 response，检查 `Bind` 错误。所有加锁/解锁和 Close 错误都必须被检查或记录，不使用空白标识符吞掉错误。

- [ ] **Step 3: 编译示例**

  Run: `go test ./ipc/...`
  Expected: PASS；示例包在 Linux/macOS/Windows 构建标签下均无编译错误。

### Task 6: 跨平台验证和文档自检

**Files:**
- Modify: `ipc/README.md`
- Test: `ipc/*_test.go`

- [ ] **Step 1: 运行 Windows 测试**

  Run: `go test ./ipc/... -count=1`
  Expected: PASS；Windows 命名管道 Accept 关闭、资源释放和业务往返测试全部完成。

- [ ] **Step 2: 交叉构建 Unix 平台**

  Run: `$env:GOOS='linux'; $env:GOARCH='amd64'; go build ./ipc/...; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; $env:GOOS='darwin'; go build ./ipc/...; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; Remove-Item Env:GOOS,Env:GOARCH`
  Expected: PASS；Unix 共享内存、锁与传输实现可编译。

- [ ] **Step 3: 交叉构建 Windows 平台**

  Run: `$env:GOOS='windows'; $env:GOARCH='amd64'; go build ./ipc/...; if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }; Remove-Item Env:GOOS,Env:GOARCH`
  Expected: PASS；Windows 系统调用签名与 build tags 正确。

- [ ] **Step 4: 最终检查**

  Run: `$files = Get-ChildItem ipc -Recurse -Filter *.go | ForEach-Object { $_.FullName }; gofmt -w $files; go test ./ipc/... -count=1; git diff --check`
  Expected: PASS；不修改 `ipc/xx.md`，没有未处理的测试失败或格式问题。
