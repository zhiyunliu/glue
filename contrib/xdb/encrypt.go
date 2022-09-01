package xdb

import (
	"fmt"
	"os"
	"strings"

	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/golibs/xsecurity/aes"
)

//默认数据库环境变量加密key
var SecretKey = "glue.xdb12345678"
var SecretMode = "cbc/pkcs7"

const (
	_connPrefix = "encrypt://"
)

func DecryptConn(conn string) (newConn string, err error) {
	if !strings.HasPrefix(conn, _connPrefix) {
		newConn = conn
		return
	}
	envName := global.Config.Get("app").Value("BASE_SECRET_ENV_NAME")
	if envName.String() == "" {
		err = fmt.Errorf("数据库配置为加密模式,但 app.BASE_SECRET_ENV_NAME 值为空")
		return
	}
	secretKey := os.Getenv(envName.String())

	orgKey, err := aes.Decrypt(secretKey, SecretKey, SecretMode)
	if err != nil {
		return
	}
	tmpConn := strings.TrimPrefix(conn, _connPrefix)
	newConn, err = aes.Decrypt(tmpConn, orgKey, SecretMode)
	return
}
