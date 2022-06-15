package xdb

import (
	"context"
	"runtime"
	"time"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xtypes"
)

//DB 数据库操作类
type xDB struct {
	db  internal.ISysDB
	tpl tpl.SQLTemplate
}

//NewDB 创建DB实例
func NewDB(proto string, conn string, maxOpen int, maxIdle int, maxLifeTime int) (obj xdb.IDB, err error) {
	if maxOpen <= 0 {
		maxOpen = runtime.NumCPU() * 10
	}
	if maxIdle <= 0 {
		maxIdle = maxOpen
	}
	if maxLifeTime <= 0 {
		maxLifeTime = 600 //10分钟
	}
	dbobj := &xDB{}
	dbobj.tpl, err = tpl.GetDBTemplate(proto)
	if err != nil {
		return
	}
	dbobj.db, err = internal.NewSysDB(proto, conn, maxOpen, maxIdle, time.Duration(maxLifeTime)*time.Second)
	return dbobj, err
}

//Query 查询数据
func (db *xDB) Query(ctx context.Context, sql string, input map[string]interface{}) (rows xdb.Rows, err error) {
	query, args := db.tpl.GetSQLContext(sql, input)
	data, err := db.db.Query(query, args...)
	if err != nil {
		return nil, getError(err, query, args)
	}
	defer func() {
		if data != nil {
			data.Close()
		}
	}()
	rows, err = resolveRows(data)
	if err != nil {
		return nil, getError(err, query, args)
	}
	return
}

//Multi 查询数据(多个数据集)
func (db *xDB) Multi(ctx context.Context, sql string, input map[string]interface{}) (datasetRows []xdb.Rows, err error) {
	query, args := db.tpl.GetSQLContext(sql, input)
	sqlRows, err := db.db.Query(query, args...)
	if err != nil {
		return nil, getError(err, query, args)
	}
	defer func() {
		if sqlRows != nil {
			sqlRows.Close()
		}
	}()
	datasetRows, err = resolveMultiRows(sqlRows)
	if err != nil {
		return nil, getError(err, query, args)
	}
	return
}

func (db *xDB) First(ctx context.Context, sql string, input map[string]interface{}) (data xdb.Row, err error) {
	rows, err := db.Query(ctx, sql, input)
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

func (db *xDB) Scalar(ctx context.Context, sql string, input map[string]interface{}) (data interface{}, err error) {
	rows, err := db.Query(ctx, sql, input)
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
func (db *xDB) Exec(ctx context.Context, sql string, input map[string]interface{}) (r xdb.Result, err error) {
	query, args := db.tpl.GetSQLContext(sql, input)
	r, err = db.db.Exec(query, args...)
	if err != nil {
		return nil, getError(err, query, args)
	}
	return
}

//Begin 创建事务
func (db *xDB) Begin() (t xdb.ITrans, err error) {
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
