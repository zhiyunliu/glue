package xipc

import "errors"

var (
	// ErrEndpointInUse 表示本地端点已被服务占用。
	ErrEndpointInUse = errors.New("ipc: endpoint already in use")
	// ErrDeadlineUnsupported 表示当前传输不支持设置 deadline。
	ErrDeadlineUnsupported = errors.New("ipc: deadlines are unsupported by this transport")
	// ErrClosed 表示客户端已关闭。
	ErrClosed = errors.New("ipc: client closed")
)

// SharedRegion 表示一块跨进程共享内存。
type SharedRegion interface {
	Bytes() []byte
	Close() error
}

// ProcessLock 表示一把跨进程互斥锁。Windows 上 Lock 和 Unlock 必须由同一 goroutine 顺序调用。
type ProcessLock interface {
	Lock() error
	Unlock() error
	Close() error
}
