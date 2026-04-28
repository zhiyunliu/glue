# Glue 框架能力与使用约束（基于代码事实）

> 本文件是 glue 框架仓库级规则文件，约束范围是框架实现、适配器扩展、示例代码和 AI 生成内容。
>
> 本文件不描述下游业务服务脚手架，不定义仓库中不存在的抽象。

## 1. 适用范围与总则

| 规则ID | 规则 |
|---|---|
| DOC-001 | 以仓库真实代码为准，优先锚定 `micro.app.go`、`cli/`、`transport/`、`engine/`、`config/`、`standard.go`、`contrib/`。 |
| DOC-002 | `glue` 根包负责统一入口与标准组件导出；`contrib/` 是适配器实现层，不是唯一能力承载层。 |
| DOC-003 | 禁止在规则、示例或生成代码中编造 `server/base.Server`、`*.ifc.go`、`OnInit/OnStart/OnRunning/OnStop`、`errors/codes.yaml` 等仓库中不存在的抽象。 |
| DOC-004 | 面向业务项目骨架的模板规范必须独立维护，不能混入本文件。 |

## 2. 核心架构地图

| 层次 | 主要目录 | 责任 |
|---|---|---|
| 应用装配 | `micro.app.go`、`cli/` | 应用入口、CLI、服务生命周期、系统服务模式。 |
| 传输契约 | `transport/` | `Server`、`Endpointer`、endpoint 与注册发现之间的统一契约。 |
| 内建服务 | `server/api`、`server/rpc`、`server/cron`、`server/mqc` | 四类内建 server 的配置、启动、路由挂载与停止。 |
| 路由执行链 | `engine/`、`router/`、`middleware/`、`context/` | 引擎适配、反射注册、统一 handler 签名、中间件链和上下文模型。 |
| 配置与组件 | `config/`、`standard/`、`container/`、`cache/`、`xdb/`、`xhttp/`、`xrpc/`、`queue/`、`dlocker/` | 配置加载、多源合并、标准组件构建与缓存。 |
| 观测与治理 | `log/`、`errors/`、`metrics/`、`opentelemetry/`、`registry/`、`selector/` | 日志、错误、指标、追踪、服务注册与选路。 |
| 外部适配 | `contrib/` | 各类协议、注册中心、配置中心、引擎和客户端实现。 |

## 3. 应用装配与生命周期

### 3.1 实际启动链路

1. `glue.NewApp` 创建 `MicroApp`，内部持有 `cli.App`。
2. `MicroApp.Start` 创建 `global.Ctx` 后委托 `cli.App.Start`。
3. `cli.App.Start` 初始化日志、装载命令并进入 `urfave/cli` 命令系统。
4. `ServiceApp.initApp` 读取命令行配置文件，构建 `config.Config`，再加载 `app` 节点到运行时设置。
5. `ServiceApp.Start` 顺序执行 `loadRegistry -> loadConfig -> apprun`。
6. `apprun` 顺序执行 `StartingHooks -> global.StartRunning -> InitOtel -> 每个 server 执行 Config/Start -> startTraceServer -> register -> StartedHooks`。
7. `buildInstance` 仅收集同时实现 `transport.Endpointer` 且 `ServiceName()` 非空的 server 进入注册信息。
8. `Stop` 顺序执行 `deregister -> StopingHooks -> 逐个 server.Stop -> StopedHooks -> closeLogger`。

### 3.2 生命周期规则

| 规则ID | 规则 |
|---|---|
| LIFE-001 | 框架真实生命周期由 `ServiceApp`、hook 和 `transport.Server` 组成，不存在统一强制实现的四段式生命周期接口。 |
| LIFE-002 | 当前实现要求 `cmdConfigFile` 可用，本地配置文件是启动主入口，远程配置源是后续追加加载，不是替代入口。 |
| LIFE-003 | 服务是否进入注册中心，取决于是否配置 `Registrar`，以及 server 是否实现 `transport.Endpointer`。 |
| LIFE-004 | 当前停止链路没有统一的“默认 5 秒 OnStop 超时”约束；注册中心有单独的 `RegistrarTimeout`，server 停止走直接调用。 |

## 4. Server 与 transport 契约

### 4.1 统一契约

所有内建或自定义 server 的最小契约都是：

```go
type Server interface {
    Name() string
    Type() string
    Start(context.Context) error
    Stop(context.Context) error
    Config(cfg config.Config) error
}
```

如果一个 server 还要参与注册与发现，则还需要实现：

```go
type Endpointer interface {
    ServiceName() string
    Endpoint() *url.URL
    RouterPathList() RouterList
}
```

### 4.2 内建能力矩阵

| 类型 | 包 | 主要能力 | 备注 |
|---|---|---|---|
| API | `server/api` | HTTP 服务、路由分组、中间件、静态资源 | `Static` 和 `StaticFile` 仅 API server 提供。 |
| RPC | `server/rpc` | RPC 服务、统一路由、编码解码、注册 endpoint | 依赖 `xrpc` 底座。 |
| CRON | `server/cron` | 定时任务 server、cron endpoint、统一 handler | 依赖 `xcron` 底座。 |
| MQC | `server/mqc` | 消息消费 server、mqc endpoint、统一 handler | 依赖 `xmqc` 底座。 |
| Custom | 任意包 | 自定义传输能力 | 至少实现 `transport.Server`；如需注册再实现 `transport.Endpointer`。 |

### 4.3 Server 规则

| 规则ID | 规则 |
|---|---|
| SRV-001 | 框架内建 server 与自定义 server 都通过 `glue.Server(...)` 接入应用，不存在必须继承的基类。 |
| SRV-002 | API、RPC、CRON、MQC 的共同路由入口是 `Use`、`Group`、`Handle`，不要为每种 server 发明单独的注册 DSL。 |
| SRV-003 | 文档或示例不得假定服务 ID 必须是 `IP:port`；当前实例 ID 默认由会话 ID 生成，endpoint 才携带地址。 |

## 5. 引擎、路由与 Handler 约束

### 5.1 engine 是适配层，不是单一实现

`engine` 包定义的是 `AdapterEngine` 与 `Resolver` 注册体系，具体实现通过 `engine.Register` 注册，再由 `engine.NewEngine` 按协议解析。默认接入依靠空导入触发 `init`，当前仓库内建了 `gin` 和 `alloter` 两类引擎适配器。

### 5.2 路由注册入口

`RouterGroup` 暴露的核心能力如下：

1. `Use(middlewares...)` 给当前组追加中间件。
2. `Group(relativePath, middlewares...)` 创建子路由组。
3. `Handle(relativePath, handler, opts...)` 挂载 handler 或对象。

### 5.3 反射注册规则

| 规则ID | 规则 |
|---|---|
| ROUTE-001 | 允许直接注册函数，函数签名必须是 `func(context.Context) interface{}`。 |
| ROUTE-002 | 允许注册对象；对象上的可识别方法后缀只有 `Handling`、`Handle`、`Handled`。 |
| ROUTE-003 | 路由合法性的硬约束是目标路径上必须存在 `Handle`；`Handling` 和 `Handled` 是可选钩子，不是强制三级链。 |
| ROUTE-004 | 反射依赖 Go 导出方法，因此源码方法名应写成 `ProfileHandle` 这类导出形式，而不是 `profileHandle`。路由子路径会由框架把前缀转成小写。 |
| ROUTE-005 | 不要在规则或示例中使用仓库中不存在的泛化写法，例如 `engine.Use(middleware.Auth(), middleware.RateLimit(...))` 这种全局伪 API。真实入口是具体 server 或 group 的 `Use`。 |

### 5.4 默认执行链

请求进入执行链后，框架默认会先挂上：

1. `otels.Server()`
2. `recovery.Recovery()`
3. group 上声明的中间件链
4. `Handling -> Handle -> Handled` 相关处理
5. 统一响应写回与请求/响应日志输出

## 6. 统一上下文与 middleware 规则

### 6.1 context 能力边界

`context.Context` 暴露的是统一上下文接口，而不是某个具体引擎对象。可稳定依赖的核心能力包括：

1. `ServerType()`、`ServerName()`
2. `Request()`、`Response()`
3. `Bind(interface{}) error`
4. `Meta()`
5. `Log()`
6. `Context()` 与 `ResetContext(...)`

其中：

1. `Request` 提供 method、client ip、request id、header、path、query、body 等读取能力。
2. `Response` 提供状态码、header、写回、重定向、flush 等能力。
3. `context/session.go` 与 `session/` 当前只提供 sid 透传辅助，不是完整用户会话系统。

### 6.2 middleware 规则

| 规则ID | 规则 |
|---|---|
| MID-001 | middleware 是跨传输统一抽象，签名是 `func(Handler) Handler`。 |
| MID-002 | middleware 的解析依赖 `MiddlewareBuilder` 注册表，运行时通过 `Resolve` 或 `BuildMiddlewareList` 从配置构建。 |
| MID-003 | 引擎启动时会默认注册 `jwt` 与 `ratelimit` 的 builder；服务端默认链还会附加 `otels` 与 `recovery`。 |
| MID-004 | 不要把 middleware 写成只能服务于 HTTP 的专用约定；框架的目标是 API、RPC、CRON、MQC 复用同一套抽象。 |

## 7. 配置模型与加载机制

### 7.1 config 抽象

`config.Config` 的核心能力是：

1. `Load()`
2. `Source(sources...)`
3. `ScanTo(v)`
4. `Value(key)`
5. `Get(key)`
6. `Watch(key, observer)`
7. `Close()`

配置对象支持多源合并、值缓存和 watcher 驱动更新。对单个节点的推荐访问方式是：

1. 用 `Get(key)` 获取子树配置。
2. 用 `Value(key)` 获取值对象。
3. 用 `ScanTo` 把结构化配置映射到结构体。

### 7.2 当前装载顺序

1. 命令行配置文件通过 `config/file` 作为本地 source 先加载。
2. 读取根配置中的 `app` 节点，设置运行模式与本地 IP 掩码。
3. 如果存在 `registry` 配置，则装载注册中心。
4. 如果存在 `config` 配置，则解析远程配置 source 并追加到当前 config。

### 7.3 配置规则

| 规则ID | 规则 |
|---|---|
| CFG-001 | 文档和代码要围绕树形配置模型书写，不要强行假定统一的 `env`、`log_level`、`db.host` 扁平字段模板。 |
| CFG-002 | 当前 README 中可稳定看到的根配置节点包括 `app`、`registry`、`config`、`caches`、`queues`、`dbs`、`servers` 以及各 contrib 对应节点。 |
| CFG-003 | `app.mode` 的实际语义是 `debug/release`，不要写成 `dev/test/prod` 强约束。 |
| CFG-004 | 不要把当前实现描述成固定的“环境变量 -> Consul -> Nacos -> 本地文件”优先级链；代码事实是本地文件先装载，远程 source 后续追加。 |

## 8. 标准组件访问模式

### 8.1 统一模式

标准组件遵循相同模型：

1. 根包在 `init` 中向 `standard` 注册 builder。
2. `standard.GetInstance(name)` 懒创建标准对象。
3. `container.Container` 负责按 `type + name + keys` 缓存真实实例。
4. 实例创建依赖 `global.Config` 当前配置。

### 8.2 对外统一入口

| 入口 | 作用 |
|---|---|
| `glue.DB(name, opts...)` | 获取数据库抽象实例。 |
| `glue.Cache(name, opts...)` | 获取缓存抽象实例。 |
| `glue.Queue(name, opts...)` | 获取队列抽象实例。 |
| `glue.RPC(name)` | 获取 RPC 客户端实例。 |
| `glue.Http(name)` | 获取 HTTP 客户端实例。 |
| `glue.DLocker(key, opts...)` | 获取分布式锁实例。 |
| `glue.Custom(name)` | 获取自定义标准对象实例。 |

### 8.3 组件规则

| 规则ID | 规则 |
|---|---|
| STD-001 | DB、Cache、Queue、Http、RPC、DLocker 应被视为统一模式下的标准组件，不要把它们描述成互不相关的工具函数集合。 |
| STD-002 | 新增基础设施能力时，优先遵循“抽象包接口 + standard builder + container 缓存 + contrib 适配器”的现有模式。 |
| STD-003 | 文档和示例不能跳过配置驱动，直接把底层实现写死到业务代码。 |

## 9. 错误、日志与观测规则

### 9.1 errors 的真实模型

| 规则ID | 规则 |
|---|---|
| ERR-001 | 标准错误模型是 `errors.Error` 接口与 `xError` 实现，核心字段是 `code`、`sub_code`、`message`。 |
| ERR-002 | 推荐通过 `errors.New`、`errors.Clone` 或 typed helper 创建错误。 |
| ERR-003 | `sub_code` 是可选的机器可读分类字段；不能把“所有错误都必须有 subcode”写成框架事实。 |
| ERR-004 | 普通 `error` 会被默认编码链兜底成 `500`，因此“禁止返回原生 error”不是当前仓库事实。 |
| ERR-005 | `WithStatusCode` 字段存在，但默认错误编码链按 `GetCode()` 写响应码，不能把它描述成默认生效机制。 |

### 9.2 日志与请求链

1. 日志基础能力来自 `log/` 对 `xlog` 的封装。
2. 请求级 logger 由引擎适配层创建并写回 context。
3. `X-Request-ID`、sid 与 logger 的 session id 在当前实现里是一条传播链。
4. 请求与响应日志主要在 `engine` 统一输出，不需要在每个 handler 手工重复接线。

### 9.3 metrics 与 opentelemetry

1. `ServiceApp.apprun` 会在 server 启动前调用 `opentelemetry.InitOtel`。
2. 服务端默认链包含 `otels.Server()`，会输出请求计数、处理中数与延迟等指标。
3. `metrics` 采用 provider 注册模式，当前默认通过空导入启用 Prometheus provider。

## 10. 服务注册、选路与客户端规则

### 10.1 registry

1. `registry` 负责服务注册、注销、查询与 watch。
2. `ServiceInstance` 由 `ServiceApp.buildInstance` 生成，元数据来自 `options.Metadata` 与构建信息。
3. 当前仓库已实现的注册中心适配器至少包括 `nacos` 与 `consul`。

### 10.2 selector 与客户端行为

| 规则ID | 规则 |
|---|---|
| DISC-001 | `selector` 是客户端选点抽象，不是注册中心本身。 |
| DISC-002 | 当前内建 selector 至少包括 `random`、`p2c`、`wrr`。 |
| DISC-003 | `xhttp` 的 `http` 适配器会结合 `registry + selector` 做服务发现与选点。 |
| DISC-004 | `xrpc` 的 `grpc` 适配器使用的是 `registry + gRPC resolver/roundrobin`，不要错误描述成默认走 `selector`。 |
| DISC-005 | 不要在文档中强加 `group`、`cluster`、实例 ID 的命名规则，除非具体适配器实现中已有校验。 |

## 11. contrib 扩展机制

### 11.1 通用模式

新增适配器时，优先遵循以下模式：

1. 在主包中定义抽象接口、resolver 或 factory 入口。
2. 在 `contrib/` 下实现具体适配器。
3. 通过 `init` 完成注册。
4. 由上层通过空导入激活。

### 11.2 当前已覆盖的主要扩展面

| 领域 | 已实现适配器 |
|---|---|
| engine | `gin`、`alloter` |
| config | `nacos`、`consul` |
| cache | `redis` |
| dlocker | `redis` |
| metrics | `prometheus` |
| registry | `nacos`、`consul` |
| xhttp | `http` |
| xrpc | `grpc` |
| xcron | `robfigcron` |
| xmqc | `alloter` |

### 11.3 扩展规则

| 规则ID | 规则 |
|---|---|
| EXT-001 | 新扩展应复用现有注册模式，不要绕过 factory 或 resolver 直接把实现写死进调用方。 |
| EXT-002 | 空导入是当前扩展激活的重要约定，文档与示例应明确这一点。 |
| EXT-003 | `contrib/` 负责适配实现，但不等于“所有能力只能在 contrib 中实现”；核心抽象和 builder 仍位于主包。 |

## 12. CLI 与运行方式

框架入口是 `urfave/cli` 与 `kardianos/service` 的组合，当前命令面至少包括：

1. `run`
2. `start`
3. `stop`
4. `restart`
5. `install`
6. `remove`
7. `status`

规则如下：

| 规则ID | 规则 |
|---|---|
| CLI-001 | 前台运行与系统服务运行共享同一个 `ServiceApp` 生命周期，不应维护两套启动逻辑。 |
| CLI-002 | 命令层先通过 `getService` 构造 `AppService`，再把控制权交给系统服务层。 |
| CLI-003 | 规则文档必须承认配置文件是 CLI 运行前提，而不是可有可无的示例参数。 |

## 13. 明确禁止的失真写法

以下说法不得再出现在规则、示例或生成代码中：

1. 服务必须继承 `server/base.Server`。
2. 所有服务必须实现 `OnInit/OnStart/OnRunning/OnStop`。
3. `Handling/Handle/Handled` 是框架强制三级链。
4. 源码方法名必须写成 `profileHandle` 这类小驼峰非导出方法。
5. 中间件统一通过不存在的全局 `engine.Use(...)` 注册。
6. 所有错误都必须在 `errors/codes.yaml` 预注册。
7. 任何原生 `error` 都禁止返回。
8. `WithStatusCode` 会自动改变默认 HTTP 错误响应状态码。
9. 配置模型固定要求 `env/log_level` 字段，且 `env` 只能是 `dev/test/prod`。
10. 配置加载顺序固定为“环境变量 -> Consul -> Nacos -> 本地文件”。
11. 服务实例 ID 必须是 `IP:port`。
12. `xrpc` 默认通过 `selector` 进行负载均衡。
13. `session` 是完整用户会话和登录态子系统。

## 14. 版本记录

| 版本 | 日期 | 说明 |
|---|---|---|
| 3.0 | 2026-04-27 | 基于仓库真实实现重写为框架级能力与使用约束，移除失真的业务模板化规范。 |
| 2.0 | 2026-04-27 | 旧版模板化规则，已废弃。 |