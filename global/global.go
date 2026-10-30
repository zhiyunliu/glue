package global

import (
	"context"

	"github.com/zhiyunliu/glue/config"
)

var (
	Mode    string = "debug"
	AppName string = ""
	// IPCSocketDir 是 Unix socket 的共享目录，可通过编译参数 -X 注入；Windows 不使用此变量。
	IPCSocketDir string = "/tmp/glue-ipcsocket"
)

var (
	Config config.Config
)

var (
	//是否包含api服务,默认false
	HasApi bool = false
)

var (
	Ctx context.Context
)
