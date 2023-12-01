package xdb

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

// DB 数据库操作类
type xDB struct {
	cfg *Setting
	db  internal.ISysDB
	tpl tpl.SQLTemplate
}

// NewDB 创建DB实例
func NewDB(proto string, setting *Setting, opts ...xdb.Option) (obj xdb.IDB, err error) {
	newCfg, err := xdb.DefaultRefactor(setting.ConnName, setting.Cfg)
	if err != nil {
		return
	}
	if newCfg != nil {
		setting.Cfg = newCfg
	}

	for i := range opts {
		opts[i](setting.Cfg)
	}

	conn := setting.Cfg.Conn
	maxOpen := setting.Cfg.MaxOpen
	maxIdle := setting.Cfg.MaxIdle
	maxLifeTime := setting.Cfg.LifeTime

	if maxOpen <= 0 {
		maxOpen = runtime.NumCPU() * 10
	}
	if maxIdle <= 0 {
		maxIdle = maxOpen
	}

	dbobj := &xDB{
		cfg: setting,
	}

	setting.slowThreshold = time.Duration(setting.Cfg.LongQueryTime) * time.Millisecond
	if setting.Cfg.LoggerName != "" {
		setting.logger, _ = xdb.GetLogger(setting.Cfg.LoggerName)
	}

	dbobj.tpl, err = tpl.GetDBTemplate(proto)
	if err != nil {
		return
	}
	dbobj.db, err = internal.NewSysDB(proto, conn, maxOpen, maxIdle, time.Duration(maxLifeTime)*time.Second)
	return dbobj, err
}
func (db *xDB) GetImpl() interface{} {
	return db.db
}

// Query 查询数据
func (db *xDB) Query(ctx context.Context, sqls string, input any) (rows xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return internal.ResolveRows(r)
	})
	rows = tmp.(xdb.Rows)
	return
}

// Multi 查询数据(多个数据集)
func (db *xDB) Multi(ctx context.Context, sqls string, input any) (datasetRows []xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return internal.ResolveMultiRows(r)
	})
	datasetRows = tmp.([]xdb.Rows)
	return
}

func (db *xDB) First(ctx context.Context, sqls string, input any) (data xdb.Row, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return internal.ResolveFirstRow(r)
	})
	data = tmp.(xdb.Row)
	return
}

func (db *xDB) Scalar(ctx context.Context, sqls string, input any) (data interface{}, err error) {
	data, err = db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return internal.ResolveScalar(r)
	})
	return
}

// Execute 根据包含@名称占位符的语句执行查询语句
func (db *xDB) Exec(ctx context.Context, sql string, input any) (r xdb.Result, err error) {

	dbParam, err := internal.ResolveParams(input)
	if err != nil {
		return
	}
	query, execArgs := db.tpl.GetSQLContext(sql, dbParam)

	start := time.Now()
	debugPrint(ctx, db.cfg, query, execArgs...)
	r, err = db.db.Exec(query, execArgs...)
	if err != nil {
		return r, internal.GetError(err, query, execArgs...)
	}
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)

	return
}

// Query 查询数据
func (db *xDB) QueryAs(ctx context.Context, sqls string, input any, results any) (err error) {
	return db.dbQueryAs(ctx, sqls, input, results, func(r *sql.Rows, a any) error {
		return internal.ResolveDataResult(r, results)
	})
}

func (db *xDB) FirstAs(ctx context.Context, sqls string, input any, result any) (err error) {
	return db.dbQueryAs(ctx, sqls, input, result, func(r *sql.Rows, a any) error {
		return internal.ResolveDataResult(r, result)
	})
}

// Begin 创建事务
func (db *xDB) Begin() (t xdb.ITrans, err error) {
	tt := &xTrans{
		cfg: db.cfg,
	}
	tt.tx, err = db.db.Begin()
	if err != nil {
		return
	}
	tt.tpl = db.tpl
	return tt, nil
}

// Transaction 执行事务
func (db *xDB) Transaction(callback xdb.TransactionCallback) (err error) {
	tt := &xTrans{
		cfg: db.cfg,
	}
	tt.tx, err = db.db.Begin()
	if err != nil {
		return
	}
	tt.tpl = db.tpl
	defer func() {
		if robj := recover(); robj != nil {
			tt.Rollback()
			rerr, ok := robj.(error)
			if !ok {
				rerr = fmt.Errorf("%+v", robj)
			}
			buf := make([]byte, 64<<10) //nolint:gomnd
			n := runtime.Stack(buf, false)
			buf = buf[:n]
			err = xdb.NewPanicError(rerr, string(buf))
		}
	}()
	err = callback(tt)
	if err != nil {
		tt.Rollback()
		return
	}
	tt.Commit()
	return
}

// Close  关闭当前数据库连接
func (db *xDB) Close() error {
	return db.db.Close()
}

func (db *xDB) dbQuery(ctx context.Context, sql string, input any, callback internal.DbResolveMapValCallback) (result any, err error) {
	dbParams, err := internal.ResolveParams(input)
	if err != nil {
		return
	}

	query, execArgs := db.tpl.GetSQLContext(sql, dbParams)

	start := time.Now()

	debugPrint(ctx, db.cfg, query, execArgs...)
	rows, err := db.db.Query(query, execArgs...)
	if err != nil {
		return nil, internal.GetError(err, query, execArgs...)
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)
	result, err = callback(rows)
	return
}

func (db *xDB) dbQueryAs(ctx context.Context, sql string, input any, result any, callback internal.DbResolveResultCallback) (err error) {
	dbParams, err := internal.ResolveParams(input)
	if err != nil {
		return
	}

	query, execArgs := db.tpl.GetSQLContext(sql, dbParams)

	start := time.Now()

	debugPrint(ctx, db.cfg, query, execArgs...)
	rows, err := db.db.Query(query, execArgs...)
	if err != nil {
		return internal.GetError(err, query, execArgs...)
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)
	err = callback(rows, result)
	return
}
