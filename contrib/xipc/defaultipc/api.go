package defaultipc

import (
	"net"

	shared "github.com/zhiyunliu/glue/xipc"
)

type SharedRegion = shared.SharedRegion
type ProcessLock = shared.ProcessLock

var ErrClosed = shared.ErrClosed
var ErrEndpointInUse = shared.ErrEndpointInUse
var ErrDeadlineUnsupported = shared.ErrDeadlineUnsupported

// Listen 创建默认协议本地监听器。
func Listen(e string) (net.Listener, error) { return listen(e) }

// Connect 建立默认协议本地连接。
func Connect(e string) (net.Conn, error) { return dialTransport(e) }

// OpenShared 打开或创建命名共享内存。
func OpenShared(n string, s int, c bool) (SharedRegion, error) { return openShared(n, s, c) }

// OpenLock 打开或创建跨进程互斥锁。
func OpenLock(n string) (ProcessLock, error) { return openLock(n) }
