package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/zhiyunliu/glue/server/ipc"
)

type updateValueArgs struct {
	Value int64 `json:"value"`
}

type valueResponse struct {
	Value int64 `json:"value"`
}

func main() {
	if err := run(); err != nil {
		log.Printf("IPC 服务退出：%v", err)
	}
}

func run() (runErr error) {
	sharedName, lockName, endpoint := exampleNames()
	region, err := ipc.OpenShared(sharedName, 16, true)
	if err != nil {
		return fmt.Errorf("打开共享内存：%w", err)
	}
	defer func() {
		if err := region.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("关闭共享内存：%w", err))
		}
	}()

	lock, err := ipc.OpenLock(lockName)
	if err != nil {
		return fmt.Errorf("打开进程锁：%w", err)
	}
	defer func() {
		if err := lock.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("关闭进程锁：%w", err))
		}
	}()

	server := ipc.New("ipc", ipc.WithAddr(endpoint))
	defer func() {
		if err := server.Close(); err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("关闭 IPC 服务：%w", err))
		}
	}()

	server.RegisterFunc("UpdateValue", func(req ipc.Request) *ipc.Response {
		var args updateValueArgs
		if err := req.Bind(&args); err != nil {
			return ipc.WrapError(ipc.CodeBadRequest, "bad_args", err)
		}
		if err := withLock(lock, func() {
			sequence := (*uint64)(unsafe.Pointer(&region.Bytes()[0]))
			atomic.AddUint64(sequence, 1)
			binary.LittleEndian.PutUint64(region.Bytes()[8:16], uint64(args.Value))
			atomic.AddUint64(sequence, 1)
		}); err != nil {
			return ipc.WrapError(ipc.CodeHandlerError, "lock_failed", err)
		}
		return ipc.OK(valueResponse{Value: args.Value})
	})

	server.RegisterFunc("Reset", func(ipc.Request) *ipc.Response {
		if err := withLock(lock, func() {
			sequence := (*uint64)(unsafe.Pointer(&region.Bytes()[0]))
			atomic.AddUint64(sequence, 1)
			binary.LittleEndian.PutUint64(region.Bytes()[8:16], 0)
			atomic.AddUint64(sequence, 1)
		}); err != nil {
			return ipc.WrapError(ipc.CodeHandlerError, "lock_failed", err)
		}
		return ipc.OK(nil)
	})

	server.RegisterFunc("Get", func(ipc.Request) *ipc.Response {
		var value int64
		if err := withLock(lock, func() {
			value = int64(binary.LittleEndian.Uint64(region.Bytes()[8:16]))
		}); err != nil {
			return ipc.WrapError(ipc.CodeHandlerError, "lock_failed", err)
		}
		return ipc.OK(valueResponse{Value: value})
	})

	log.Printf("IPC 服务监听端点 %q", endpoint)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.Start(ctx); err != nil {
		return err
	}
	<-ctx.Done()
	return server.Stop(context.Background())
}

func withLock(lock ipc.ProcessLock, action func()) (err error) {
	if err := lock.Lock(); err != nil {
		return err
	}
	defer func() {
		if unlockErr := lock.Unlock(); unlockErr != nil {
			err = errors.Join(err, unlockErr)
		}
	}()
	action()
	return nil
}

func exampleNames() (sharedName, lockName, endpoint string) {
	if runtime.GOOS == "windows" {
		return `Local\glue_ipc_example`, `Local\glue_ipc_example_lock`, "glue-ipc-example"
	}
	return "/glue_ipc_example",
		filepath.Join(os.TempDir(), "glue-ipc-example.lock"),
		filepath.Join(os.TempDir(), "glue-ipc-example.sock")
}
