//go:build linux || darwin

package defaultipc

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/zhiyunliu/glue/global"
)

func TestUnixSocketAcrossWorkingDirectories(t *testing.T) {
	if dir := os.Getenv("GLUE_IPC_TEST_DIR"); dir != "" {
		global.IPCSocketDir = dir
	} else {
		previous := global.IPCSocketDir
		global.IPCSocketDir = filepath.Join(t.TempDir(), "ipc")
		t.Cleanup(func() { global.IPCSocketDir = previous })
	}
	if endpoint := os.Getenv("GLUE_IPC_TEST_ENDPOINT"); endpoint != "" {
		conn, err := dialTransport(endpoint)
		if err != nil {
			t.Fatal(err)
		}
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
		return
	}

	t.Chdir(t.TempDir())
	endpoint := "glue-" + strconv.Itoa(os.Getpid())
	listener, err := listen(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	if err := listener.(*net.UnixListener).SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		t.Fatal(err)
	}

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(executable, "-test.run=^TestUnixSocketAcrossWorkingDirectories$")
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "GLUE_IPC_TEST_ENDPOINT="+endpoint, "GLUE_IPC_TEST_DIR="+global.IPCSocketDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("不同工作目录的客户端连接失败: %v\n%s", err, output)
	}
	conn, err := listener.Accept()
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestResolveUnixEndpoint(t *testing.T) {
	previous := global.IPCSocketDir
	t.Cleanup(func() { global.IPCSocketDir = previous })
	dir := t.TempDir()
	for _, tt := range []struct {
		name, dir, endpoint, want string
		wantErr                   bool
	}{
		{"共享目录", dir, "glue-test", filepath.Join(dir, "glue-test"), false},
		{"绝对端点", "relative", filepath.Join(dir, "explicit"), filepath.Join(dir, "explicit"), false},
		{"空名称", dir, "", "", true},
		{"空目录", "", "glue-test", "", true},
		{"相对目录", "relative", "glue-test", "", true},
		{"越界路径", dir, "../glue-test", "", true},
		{"当前目录", dir, ".", "", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			global.IPCSocketDir = tt.dir
			got, err := resolveUnixEndpoint(tt.endpoint)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("got %q, %v; want %q, error=%v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

func TestPrepareSocketDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "ipc")
	if err := prepareSocketDir(dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("目录权限为 %o", info.Mode().Perm())
	}
	if err := os.Chmod(dir, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := prepareSocketDir(dir); err == nil {
		t.Fatal("应拒绝其他用户可写目录")
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
	if err := prepareSocketDir(link); err == nil {
		t.Fatal("应拒绝符号链接目录")
	}
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := prepareSocketDir(file); err == nil {
		t.Fatal("应拒绝普通文件")
	}
}

func TestUnixDialDoesNotCreateDirectory(t *testing.T) {
	previous := global.IPCSocketDir
	t.Cleanup(func() { global.IPCSocketDir = previous })
	global.IPCSocketDir = filepath.Join(t.TempDir(), "missing")
	if conn, err := dialTransport("glue-test"); err == nil {
		conn.Close()
		t.Fatal("不存在的 socket 不应连接成功")
	}
	if _, err := os.Stat(global.IPCSocketDir); !os.IsNotExist(err) {
		t.Fatalf("客户端不应创建目录: %v", err)
	}
}

func TestUnixDuplicateListener(t *testing.T) {
	previous := global.IPCSocketDir
	t.Cleanup(func() { global.IPCSocketDir = previous })
	global.IPCSocketDir = filepath.Join(t.TempDir(), "ipc")
	listener, err := listen("glue-test")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	other, err := listen("glue-test")
	if other != nil {
		other.Close()
	}
	if !errors.Is(err, ErrEndpointInUse) {
		t.Fatalf("重复监听错误: %v", err)
	}
}
