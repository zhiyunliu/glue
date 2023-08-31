package internal

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zhiyunliu/golibs/xtypes"
)

var nameMap = xtypes.SMap{
	"ora":    "oci8",
	"oracle": "oci8",
	"sqlite": "sqlite3",
}

type ISysDB interface {
	Query(string, ...interface{}) (*sql.Rows, error)
	Exec(string, ...interface{}) (sql.Result, error)
	Begin() (ISysTrans, error)
	Close() error
}

// SysDB 数据库实体
type sysDB struct {
	proto       string
	conn        string
	db          *sql.DB
	maxIdle     int
	maxOpen     int
	maxLifeTime time.Duration
}

// NewSysDB 创建DB实例
func NewSysDB(proto string, conn string, maxOpen int, maxIdle int, maxLifeTime time.Duration) (ISysDB, error) {
	var err error
	if proto == "" || conn == "" {
		err = errors.New("proto or conn not allow null")
		return nil, err
	}

	obj := &sysDB{
		proto:       proto,
		conn:        conn,
		maxIdle:     maxIdle,
		maxOpen:     maxOpen,
		maxLifeTime: maxLifeTime,
	}
	proto = strings.ToLower(proto)
	proto = nameMap.GetWithDefault(proto, proto)
	obj.db, err = sql.Open(proto, conn)
	if err != nil {
		return nil, fmt.Errorf("NewSysDB.Open.proto:%s,conn=%s,error:%w", proto, conn, err)
	}
	obj.db.SetMaxIdleConns(maxIdle)
	obj.db.SetMaxOpenConns(maxOpen)
	obj.db.SetConnMaxLifetime(maxLifeTime)
	err = obj.db.Ping()
	if err != nil {
		return nil, fmt.Errorf("NewSysDB.Ping.proto:%s,conn=%s,error:%w", proto, conn, err)
	}
	return obj, nil
}

// Query 执行SQL查询语句
func (db *sysDB) Query(query string, args ...interface{}) (rows *sql.Rows, err error) {
	rows, err = db.db.Query(query, args...)
	if err != nil {
		if rows != nil {
			rows.Close()
		}
		return
	}
	return
}

// Exec 执行SQL操作语句
func (db *sysDB) Exec(query string, args ...interface{}) (result sql.Result, err error) {
	result, err = db.db.Exec(query, args...)
	if err != nil {
		return
	}
	return result, err
}

// Begin 创建一个事务请求
func (db *sysDB) Begin() (r ISysTrans, err error) {
	t := &sysTrans{}
	t.tx, err = db.db.Begin()
	return t, err
}

// Close 关闭数据库
func (db *sysDB) Close() error {
	return db.db.Close()
}
