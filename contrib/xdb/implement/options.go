package implement

type Option func(*sysDB)

func WithConnName(connName string) Option {
	return func(db *sysDB) {
		db.connName = connName
	}
}

func WithMaxOpen(maxOpen int) Option {
	return func(sd *sysDB) {
		sd.maxOpen = maxOpen
	}
}

func WithMaxIdle(maxIdle int) Option {
	return func(sd *sysDB) {
		sd.maxIdle = maxIdle
	}
}

func WithMaxLifeTime(maxLifeTime int) Option {
	return func(sd *sysDB) {
		sd.maxLifeTime = maxLifeTime
	}
}
