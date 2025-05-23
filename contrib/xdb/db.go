package xdb

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/zhiyunliu/glue/contrib/xdb/implement"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

// DB 数据库操作类
type xDB struct {
	cfg *Setting
	db  implement.ISysDB
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
	dbobj.db, err = implement.NewSysDB(proto, setting.Cfg.Conn,
		implement.WithConnName(setting.ConnName),
		implement.WithMaxOpen(setting.Cfg.MaxOpen),
		implement.WithMaxIdle(setting.Cfg.MaxIdle),
		implement.WithMaxLifeTime(setting.Cfg.LifeTime),
	)
	return dbobj, err
}
func (db *xDB) GetImpl() interface{} {
	return db.db
}

// Query 查询数据
func (db *xDB) Query(ctx context.Context, sqls string, input any) (rows xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveRows(r)
	})
	if err != nil {
		return
	}
	rows = tmp.(xdb.Rows)
	return
}

// Multi 查询数据(多个数据集)
func (db *xDB) Multi(ctx context.Context, sqls string, input any) (datasetRows []xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveMultiRows(r)
	})
	if err != nil {
		return
	}
	datasetRows = tmp.([]xdb.Rows)
	return
}

func (db *xDB) First(ctx context.Context, sqls string, input any) (data xdb.Row, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveFirstRow(r)
	})
	if err != nil {
		return
	}
	data = tmp.(xdb.Row)
	return
}

func (db *xDB) Scalar(ctx context.Context, sqls string, input any) (data interface{}, err error) {
	data, err = db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveScalar(r)
	})
	return
}

// Execute 根据包含@名称占位符的语句执行查询语句
func (db *xDB) Exec(ctx context.Context, sql string, input any) (r xdb.Result, err error) {

	dbParam, err := implement.ResolveParams(input)
	if err != nil {
		return
	}
	query, execArgs, err := db.tpl.GetSQLContext(sql, dbParam)
	if err != nil {
		err = implement.GetError(err, sql, input)
		return
	}

	start := time.Now()
	debugPrint(ctx, db.cfg, query, execArgs...)
	r, err = db.db.Exec(query, execArgs...)
	if err != nil {
		return r, implement.GetError(err, query, execArgs...)
	}
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)

	return
}

// Query 查询数据
func (db *xDB) QueryAs(ctx context.Context, sqls string, input any, results any) (err error) {
	return db.dbQueryAs(ctx, sqls, input, results, func(r *sql.Rows, val any) error {
		return implement.ResolveRowsDataResult(r, val)
	})
}

func (db *xDB) FirstAs(ctx context.Context, sqls string, input any, result any) (err error) {
	return db.dbQueryAs(ctx, sqls, input, result, func(r *sql.Rows, val any) error {
		return implement.ResolveFirstDataResult(r, val)
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

func (db *xDB) dbQuery(ctx context.Context, sql string, input any, callback implement.DbResolveMapValCallback) (result any, err error) {
	dbParams, err := implement.ResolveParams(input)
	if err != nil {
		return
	}

	query, execArgs, err := db.tpl.GetSQLContext(sql, dbParams)
	if err != nil {
		err = implement.GetError(err, sql, input)
		return
	}

	start := time.Now()

	debugPrint(ctx, db.cfg, query, execArgs...)
	rows, err := db.db.Query(query, execArgs...)
	if err != nil {
		return nil, implement.GetError(err, query, execArgs...)
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

func (db *xDB) dbQueryAs(ctx context.Context, sql string, input any, result any, callback implement.DbResolveResultCallback) (err error) {
	dbParams, err := implement.ResolveParams(input)
	if err != nil {
		return
	}

	query, execArgs, err := db.tpl.GetSQLContext(sql, dbParams)
	if err != nil {
		err = implement.GetError(err, sql, input)
		return
	}

	start := time.Now()

	debugPrint(ctx, db.cfg, query, execArgs...)
	rows, err := db.db.Query(query, execArgs...)
	if err != nil {
		return implement.GetError(err, query, execArgs...)
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
