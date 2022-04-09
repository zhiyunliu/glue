package xdb

import (
	"time"

	"github.com/zhiyunliu/gel/xdb/internal"
	"github.com/zhiyunliu/gel/xdb/tpl"
	"github.com/zhiyunliu/golibs/xtypes"
)

//IDB 数据库操作接口,安装可需能需要执行export LD_LIBRARY_PATH=/usr/local/lib
type IDB interface {
	Executer
	Begin() (ITrans, error)
	Close() error
}

//ITrans 数据库事务接口
type ITrans interface {
	Executer
	Rollback() error
	Commit() error
}

//Executer 数据库操作对象集合
type Executer interface {
	Query(sql string, input map[string]interface{}) (data Rows, err error)
	First(sql string, input map[string]interface{}) (data Row, err error)
	Scalar(sql string, input map[string]interface{}) (data interface{}, err error)
	Exec(sql string, input map[string]interface{}) (r Result, err error)
	//	ExecSp(procName string, input map[string]interface{}) (r Result, err error)
}

//DB 数据库操作类
type xDB struct {
	db  internal.ISysDB
	tpl tpl.SQLTemplate
}

//NewDB 创建DB实例
func NewDB(proto string, connString string, maxOpen int, maxIdle int, maxLifeTime int) (obj IDB, err error) {
	dbobj := &xDB{}
	dbobj.tpl, err = tpl.GetDBTemplate(proto)
	if err != nil {
		return
	}
	dbobj.db, err = internal.NewSysDB(proto, connString, maxOpen, maxIdle, time.Duration(maxLifeTime)*time.Second)
	return dbobj, err
}

//GetTPL 获取模板翻译参数
func (db *xDB) GetTPL() tpl.SQLTemplate {
	return db.tpl
}

//Query 查询数据
func (db *xDB) Query(sql string, input map[string]interface{}) (rows Rows, err error) {
	query, args := db.tpl.GetSQLContext(sql, input)
	data, err := db.db.Query(query, args...)
	if err != nil {
		return nil, getError(err, query, args)
	}
	rows, err = resolveRows(data, 0)
	if err != nil {
		return nil, getError(err, query, args)
	}
	return
}

func (db *xDB) First(sql string, input map[string]interface{}) (data Row, err error) {
	rows, err := db.Query(sql, input)
	if err != nil {
		return
	}
	if rows.IsEmpty() {
		data = make(xtypes.XMap)
		return
	}
	data = rows[0]
	return
}

func (db *xDB) Scalar(sql string, input map[string]interface{}) (data interface{}, err error) {
	rows, err := db.Query(sql, input)
	if err != nil {
		return
	}
	if rows.Len() == 0 || len(rows[0]) == 0 {
		return nil, nil
	}
	data, _ = rows[0].Get(rows[0].Keys()[0])
	return
}

//Execute 根据包含@名称占位符的语句执行查询语句
func (db *xDB) Exec(sql string, input map[string]interface{}) (r Result, err error) {
	query, args := db.tpl.GetSQLContext(sql, input)
	r, err = db.db.Exec(query, args...)
	if err != nil {
		return nil, getError(err, query, args)
	}
	return
}

//ExecuteSP 根据包含@名称占位符的语句执行查询语句
func (db *xDB) ExecSp(procName string, input map[string]interface{}, output ...interface{}) (r Result, err error) {
	query, args := db.tpl.GetSPContext(procName, input)
	ni := append(args, output...)
	r, err = db.db.Exec(query, ni...)
	if err != nil {
		return nil, getError(err, query, ni)
	}
	return
}

//Begin 创建事务
func (db *xDB) Begin() (t ITrans, err error) {
	tt := &xTrans{}
	tt.tx, err = db.db.Begin()
	if err != nil {
		return
	}
	tt.tpl = db.tpl
	return tt, nil
}

//Close  关闭当前数据库连接
func (db *xDB) Close() error {
	return db.db.Close()
}
