package global

import (
	"github.com/zhiyunliu/glue/config"
)

var (
	Mode    string = "debug"
	AppName string = ""
)

var (
	Config config.Config
)

var (
	//是否包含api服务,默认false
	HasApi bool = false
)
