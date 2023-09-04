package xdb

//数据库连接重构方法
var ConnRefactor func(connName string, cfg *Config) (newcfg *Config, err error)

func DefaultRefactor(connName string, cfg *Config) (newcfg *Config, err error) {
	newcfg = cfg
	if DecryptConn != nil {
		newcfg.Conn, err = DecryptConn(connName, cfg.Conn)
		if err != nil {
			return
		}
	}
	if ConnRefactor != nil {
		newcfg, err = ConnRefactor(connName, newcfg)
	}
	return
}
