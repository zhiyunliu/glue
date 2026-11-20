# 数据库流式查询 + SSE 推送

> 适用版本：`support_ipc_server` 分支 `dc64216` 之后（`xdb.DataRowReader` / `xdb.StreamChan`）

## 1. 解决什么问题

以往 `QueryAs` 只能把结果一次性读进一个切片，再整体返回给调用方。数据量大或查询慢时，存在两个问题：

- 服务端要把全部结果放在内存里；
- 客户端要等整个查询结束才能看到第一条数据。

现在 `QueryAs` 支持传入一个 **行读取器（`DataRowReader`）**：数据库每扫描出一行，就立刻交给读取器处理。配合 `xsse.ResultChan`，就能把查询结果**逐行通过 SSE（Server-Sent Events）推送给前端**，边查边发。

## 2. 核心对象

| 对象 | 所在包 | 作用 |
|---|---|---|
| `DataRowReader` | `glue/xdb` | 行读取器接口。`QueryAs` 发现结果参数实现了该接口，就走逐行模式。 |
| `StreamChan[T]` | `glue/xdb` | `DataRowReader` 的内置实现，把每一行（`*T`）写进一个 channel。 |
| `ResultChan[T]` | `golibs/xsse` | 从 channel 读取数据并转换为 SSE 事件，handler 直接返回它即可。 |

`DataRowReader` 接口定义：

```go
type DataRowReader interface {
	GetNewRow() any            // 为下一行创建一个新的接收对象（结构体指针或 map 指针）
	PresentRow(data any) error // 一行扫描完成后回调，返回 error 会中止查询
	Close() error              // 结果集读完（或出错）后调用
}
```

## 3. 数据流

```
 ┌──────────── 后台 goroutine ────────────┐          ┌────────── handler 所在 goroutine ──────────┐
 │                                        │          │                                            │
 │ QueryAs ──► GetNewRow() ──► 扫描一行    │          │  ResultChan ──► SSE 编码 ──► Flush ──► 客户端 │
 │                 ▲              │       │ dataChan │      ▲                                      │
 │                 └── PresentRow ┴───────┼────────► ┼──────┘                                      │
 │                                        │          │                                            │
 │ 结果集读完 ──► Close() ──► close(chan) ─┼────────► ┼─ channel 关闭 ──► SSE 结束，请求完成          │
 └────────────────────────────────────────┘          └────────────────────────────────────────────┘
```

- 查询在后台 goroutine 中执行，handler 立即返回 `ResultChan`，框架开始输出 SSE。
- channel 关闭即代表数据发送完毕，SSE 响应随之结束。
- 使用无缓冲 channel 时，数据库读取速度会被客户端接收速度"拖住"（天然背压），服务端内存占用最小。

## 4. 快速上手

### 4.1 服务端

```go
package demo

import (
	"github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xsse"
)

type dataItem struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

const querySQL = `select top 10 id, name from test order by id desc`

func StreamHandle(ctx context.Context) interface{} {
	// ① 在 handler 返回前取出需要的上下文信息，goroutine 里不要再碰 ctx
	sid := ctx.Log().SessionID()
	reqCtx := ctx.Context()

	// ② 一个 channel 同时交给"生产者"(StreamChan) 和"消费者"(ResultChan)
	dataChan := make(chan *dataItem)
	stream := xdb.NewStreamChan(dataChan)
	result := xsse.NewResultChan(reqCtx, dataChan)

	// ③ 后台执行查询
	go func() {
 		if err := glue.DB("microsql").QueryAs(reqCtx, querySQL, nil, stream); err != nil {
			log.Errorf("[%s] 流式查询失败: %v", sid, err)
		}
	}()

	// ④ 直接返回，框架识别为 SSE 响应
	return result
}
```

路由注册与普通接口一致：

```go
apiSrv := api.New("apiserver")
apiSrv.Handle("/demo/stream", StreamHandle)
```

### 4.2 客户端

命令行验证（`-N` 关闭缓冲，可以看到逐条到达）：

```bash
curl -N http://127.0.0.1:8080/demo/stream
```

输出：

```
id:1
data:{"id":10,"name":"foo"}

id:2
data:{"id":9,"name":"bar"}

```

浏览器：

```js
const es = new EventSource('/demo/stream');
es.onmessage = (e) => console.log(JSON.parse(e.data));
es.onerror   = () => es.close(); // 服务端发送完毕会关闭连接，这里避免 EventSource 自动重连
```

## 5. 使用要点

### 5.1 行类型

- `T` 必须是**结构体**或 **map**，`StreamChan` 每行都会 `new(T)` 一个新对象，各行之间互不影响。
- 结构体字段与列的映射标签优先级：`xdb` > `db` > `json`。
- SSE 的 `data` 默认使用 `json.Marshal` 序列化；`id` 默认为从 1 开始的序号。
- 如需自定义事件，可让行类型实现以下方法：

```go
func (d *dataItem) GetEventId() string   { return strconv.FormatInt(d.Id, 10) } // 自定义 id
func (d *dataItem) GetEventType() string { return "row" }                      // 自定义 event，前端用 es.addEventListener('row', ...)
```

### 5.2 必须遵守的三条

| # | 规则 | 不遵守的后果 |
|---|---|---|
| 2 | goroutine 里**不要使用 `ctx`**（包括 `ctx.Log()`、`ctx.Request()` 等） | 查询结束 → channel 关闭 → 请求结束，框架会立即回收 `ctx`（放回对象池并清空）。goroutine 稍后再访问 `ctx` 可能出现空指针 panic（导致进程退出），或串到其他请求上。请像示例那样提前取出 `sid`、`reqCtx`，日志使用包级 `log.Errorf`。 |
| 3 | 每个请求单独创建 `dataChan` / `StreamChan` / `ResultChan` | 三者都是一次性对象，channel 关闭后不可复用。 |

### 5.3 其他说明

- **中途出错对客户端不可见**：如果读到一半发生扫描错误，channel 同样会被关闭，客户端看到的是"正常结束"，无法区分数据是否完整。对完整性敏感的场景，可以在行数据中带上序号或总数，由前端自行校验。
- **连接占用**：流式输出期间数据库连接一直被占用，直到最后一行发送完。长时间、大量并发的流式查询需要关注连接池大小（`max_open`）。
- **事务中同样可用**：`ITrans.QueryAs` 走同一套逻辑。
- **只对 `QueryAs` 生效**：`FirstAs` 等其他方法不识别 `DataRowReader`。

### 5.4 已知限制

客户端中途断开时，`ResultChan` 会停止读取，但 `StreamChan.PresentRow` 是阻塞写入且不感知 `ctx`，后台 goroutine 会一直阻塞在写 channel 上而无法退出（数据库连接会因 `ctx` 取消被 `database/sql` 释放，但 goroutine 本身会泄漏）。在该问题修复前，建议：

- 只对**结果集有上限**的查询使用流式输出（如示例中的 `top N`）；
- 对外开放、客户端可能频繁中断的接口谨慎使用。

## 6. 进阶：自定义 DataRowReader

`StreamChan` 只是 `DataRowReader` 的一种实现。凡是"边查边处理"的场景都可以自己实现，例如把查询结果直接写成 CSV，不在内存中积累：

```go
type csvRowWriter struct {
	w *csv.Writer
}

var _ xdb.DataRowReader = (*csvRowWriter)(nil)

func (r *csvRowWriter) GetNewRow() any { return &dataItem{} }

func (r *csvRowWriter) PresentRow(data any) error {
	item := data.(*dataItem)
	return r.w.Write([]string{strconv.FormatInt(item.Id, 10), item.Name})
}

func (r *csvRowWriter) Close() error {
	r.w.Flush()
	return r.w.Error()
}

// 使用：同步执行即可，无需 goroutine
err := glue.DB("microsql").QueryAs(ctx.Context(), querySQL, nil, &csvRowWriter{w: csv.NewWriter(file)})
```

实现时注意：

- `GetNewRow` 必须返回**非 nil 的结构体指针或 map 指针**，否则报 `InvalidArgTypeError`。
- 如果 `PresentRow` 会把对象交给其他 goroutine 异步处理，`GetNewRow` **每次都要返回新对象**，否则后一行会覆盖前一行的值。
- `PresentRow` 返回 error 会立即中止查询，该 error 会作为 `QueryAs` 的返回值。
- `Close` 可能被调用多次（框架一次、业务兜底一次），需要保证幂等。
