package xdb

//数据库连接重构方法
var ConnRefactor func(string) (string, error)

func DefaultRefactor(conn string) (newConn string, err error) {
	newConn = conn
	if DecryptConn != nil {
		newConn, err = DecryptConn(newConn)
		if err != nil {
			return
		}
	}
	if ConnRefactor != nil {
		newConn, err = ConnRefactor(newConn)
	}
	return
}
