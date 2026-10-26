package ipc

import (
	"errors"
	"net"
)

var (
	// ErrEndpointInUse 表示本地端点已被服务占用。
	ErrEndpointInUse = errors.New("ipc: endpoint already in use")
	// ErrDeadlineUnsupported 表示当前传输不支持设置 deadline。
	ErrDeadlineUnsupported = errors.New("ipc: deadlines are unsupported by this transport")
	// ErrClosed 表示客户端已关闭。
	ErrClosed = errors.New("ipc: client closed")
	// ErrServerAlreadyServing 表示 Serve 已经在运行。
	ErrServerAlreadyServing = errors.New("ipc: server is already serving")
)

// Listen 在本地创建一个 IPC 监听器。
func Listen(endpoint string) (net.Listener, error) {
	return listen(endpoint)
}

// SharedRegion 表示一块跨进程共享内存。
type SharedRegion interface {
	Bytes() []byte
	Close() error
}

// OpenShared 打开或创建一块命名共享内存。
func OpenShared(name string, size int, create bool) (SharedRegion, error) {
	return openShared(name, size, create)
}

// ProcessLock 表示一把跨进程互斥锁。Windows 上 Lock 和 Unlock 必须由同一 goroutine 顺序调用。
type ProcessLock interface {
	Lock() error
	Unlock() error
	Close() error
}

// OpenLock 打开或创建一把命名跨进程互斥锁。
func OpenLock(name string) (ProcessLock, error) {
	return openLock(name)
}
