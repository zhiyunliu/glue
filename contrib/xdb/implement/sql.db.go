package implement

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xtypes"
)

var (
	_ xdb.DbConn = &sysDB{}
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
	connName    string
	proto       string
	conn        string
	db          *sql.DB
	maxIdle     int
	maxOpen     int
	maxLifeTime int
}

// NewSysDB 创建DB实例
func NewSysDB(proto string, conn string, opts ...Option) (ISysDB, error) {
	var err error
	if proto == "" || conn == "" {
		err = errors.New("proto or conn not allow null")
		return nil, err
	}

	obj := &sysDB{
		proto: proto,
		conn:  conn,
	}

	for i := range opts {
		opts[i](obj)
	}
	
	if obj.maxOpen <= 0 {
		obj.maxOpen = runtime.NumCPU() * 10
	}
	if obj.maxIdle <= 0 {
		obj.maxIdle = obj.maxOpen
	}
	if obj.maxLifeTime <= 0 {
		obj.maxLifeTime = 600 //默认10分钟
	}

	proto = strings.ToLower(proto)
	proto = nameMap.GetWithDefault(proto, proto)
	obj.db, err = sql.Open(proto, conn)
	if err != nil {
		return nil, fmt.Errorf("NewSysDB.Open.proto:%s,connName:%s,error:%w", proto, obj.connName, err)
	}
	obj.db.SetMaxIdleConns(obj.maxIdle)
	obj.db.SetMaxOpenConns(obj.maxOpen)
	obj.db.SetConnMaxLifetime(time.Duration(obj.maxLifeTime) * time.Second)
	err = obj.db.Ping()
	if err != nil {
		return nil, fmt.Errorf("NewSysDB.Ping.proto:%s,connName:%s,error:%w", proto, obj.connName, err)
	}
	return obj, nil
}

func (db *sysDB) GetSqlDB() *sql.DB {
	return db.db
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
