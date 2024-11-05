package xdb

import (
	"fmt"
	"os"
	"strings"

	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/golibs/xsecurity/aes"
)

var (
	SecretKey         = "glue.xdb12345678"
	SecretMode        = "cbc/pkcs7"
	BaseSecretEnvName = "BASE_SECRET_ENV_NAME"
)

const (
	_connPrefix = "encrypt://"
)

// DecryptConn 解密数据库链接
var DecryptConn func(connName, conn string) (newConn string, err error) = defaultDecryptConn

func defaultDecryptConn(connName, conn string) (newConn string, err error) {
	if !strings.HasPrefix(conn, _connPrefix) {
		newConn = conn
		return
	}
	envName := global.Config.Get("app").Value(BaseSecretEnvName)
	if envName.String() == "" {
		err = fmt.Errorf("数据库配置为加密模式,但 app.%s 值为空", BaseSecretEnvName)
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
