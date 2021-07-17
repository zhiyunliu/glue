package libs

import "crypto/md5"

func Md5(val string) (r string) {
	hash := md5.New()
	hash.Reset()
	bytes := hash.Sum([]byte(val))
	return string(bytes)
}
