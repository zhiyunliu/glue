package dbs

//DB 数据库配置
type Config struct {
	Provider   string `json:"provider" valid:"required"`
	ConnString string `json:"connString" valid:"required" label:"连接字符串"`
	MaxOpen    int    `json:"maxOpen" valid:"required" label:"最大打开连接数"`
	MaxIdle    int    `json:"maxIdle" valid:"required" label:"最大空闲连接数"`
	LifeTime   int    `json:"lifeTime" valid:"required" label:"单个连接时长(秒)"`
}

//New 构建DB连接信息
func New(provider string, connString string, opts ...Option) *Config {
	db := &Config{
		Provider:   provider,
		ConnString: connString,
		MaxOpen:    10,
		MaxIdle:    10,
		LifeTime:   600,
	}
	for _, opt := range opts {
		opt(db)
	}
	return db
}

//Option 配置选项
type Option func(*Config)

func WithMaxOpen(maxOpen int) Option {
	return func(a *Config) {
		a.MaxOpen = maxOpen
	}
}

func WithMaxIdle(maxIdle int) Option {
	return func(a *Config) {
		a.MaxIdle = maxIdle
	}
}

func WithLifeTime(lifeTime int) Option {
	return func(a *Config) {
		a.LifeTime = lifeTime
	}
}
