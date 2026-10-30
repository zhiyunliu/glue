//go:build windows

package defaultipc

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/zhiyunliu/glue/global"
)

func TestWindowsPipeIgnoresSocketDirectory(t *testing.T) {
	previous := global.IPCSocketDir
	global.IPCSocketDir = "invalid-relative-unix-directory"
	t.Cleanup(func() { global.IPCSocketDir = previous })
	endpoint := fmt.Sprintf("glue-test-%d-%d", os.Getpid(), time.Now().UnixNano())
	listener, err := listen(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			var data [1]byte
			_, err = conn.Read(data[:])
			if err == nil && data[0] != 42 {
				err = fmt.Errorf("收到错误数据: %d", data[0])
			}
			conn.Close()
		}
		accepted <- err
	}()
	conn, err := dialTransport(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.Write([]byte{42}); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-accepted:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("命名管道传输超时")
	}
}
