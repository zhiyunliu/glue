package constants

import (
	"net/http"

	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/glue/errors/subcode"
)

var ErrRemoteRequest = errors.New(http.StatusInternalServerError, "远程服务调用失败", errors.WithSubCode(subcode.IsvRemoteRequest))
