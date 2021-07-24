package libs

import (
	"crypto/md5"
	"fmt"
)

func Md5(val string) (r string) {
	return fmt.Sprintf("%x", md5.Sum([]byte(val)))
}
