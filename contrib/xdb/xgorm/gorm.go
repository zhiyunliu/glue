package xgorm

import (
	"fmt"
	"log"
	"runtime"
	"time"

	xlog "github.com/zhiyunliu/glue/log"

	contribxdb "github.com/zhiyunliu/glue/contrib/xdb"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	callbackCache = map[string]DriverOpenCallback{}
)

type DriverOpenCallback func(dsn string) gorm.Dialector

func getDriverOpenCallback(driverName string) (callback DriverOpenCallback, err error) {
	if cb, ok := callbackCache[driverName]; ok {
		callback = cb
		return
	}
	err = fmt.Errorf("没有注册[%s]的数据库驱动,请确认编译的时候是否增加对应的tag标签 go build -tags=%s", driverName, driverName)
	return
}

func buildGormDB(driverName string, cfg *contribxdb.Config) (db *gorm.DB, err error) {
	if cfg.MaxOpen <= 0 {
		cfg.MaxOpen = runtime.NumCPU() * 5
	}
	if cfg.MaxIdle <= 0 {
		cfg.MaxIdle = cfg.MaxOpen
	}
	if cfg.LifeTime <= 0 {
		cfg.LifeTime = 600 //10分钟
	}
	if cfg.LongQueryTime <= 0 {
		cfg.LongQueryTime = 500
	}

	w := log.New(xlog.DefaultLogger, "\r\n", log.LstdFlags)
	var newLogger = logger.New(
		w,
		logger.Config{
			SlowThreshold: time.Duration(cfg.LongQueryTime) * time.Millisecond, // 慢 SQL 阈值
			LogLevel:      logger.Info,                                         // Log level
			Colorful:      true,                                                // 开启彩色打印
		},
	)

	callback, err := getDriverOpenCallback(driverName)
	if err != nil {
		return
	}

	db, err = gorm.Open(callback(cfg.Conn), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.LifeTime) * time.Second)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.LifeTime) * time.Second)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	return db, nil
}

type sqlResult struct {
	lastInsertId int64
	rowsAffected int64
}

func (r sqlResult) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

func (r *sqlResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}
