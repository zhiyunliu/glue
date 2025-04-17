package xdb

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/zhiyunliu/glue/contrib/xdb/implement"
	"github.com/zhiyunliu/glue/xdb"
)

// DB 数据库操作类
type xDB struct {
	cfg   *Setting
	proto string
	db    implement.ISysDB
	tpl   xdb.SQLTemplate
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
		cfg:   setting,
		proto: setting.Cfg.Proto,
	}

	setting.slowThreshold = time.Duration(setting.Cfg.LongQueryTime) * time.Millisecond
	if setting.Cfg.LoggerName != "" {
		setting.logger, _ = xdb.GetLogger(setting.Cfg.LoggerName)
	}

	dbobj.tpl, err = xdb.GetTemplate(proto)
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
func (db *xDB) Query(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (rows xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveRows(db.proto, r)
	}, opts...)
	if err != nil {
		return
	}
	rows = tmp.(xdb.Rows)
	return
}

// Multi 查询数据(多个数据集)
func (db *xDB) Multi(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (datasetRows []xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveMultiRows(db.proto, r)
	}, opts...)
	if err != nil {
		return
	}
	datasetRows = tmp.([]xdb.Rows)
	return
}

func (db *xDB) First(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (data xdb.Row, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveFirstRow(db.proto, r)
	}, opts...)
	if err != nil {
		return
	}
	data = tmp.(xdb.Row)
	return
}

func (db *xDB) Scalar(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (data interface{}, err error) {
	data, err = db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveScalar(db.proto, r)
	}, opts...)
	return
}

// Execute 根据包含@名称占位符的语句执行查询语句
func (db *xDB) Exec(ctx context.Context, sql string, input any, opts ...xdb.TemplateOption) (r xdb.Result, err error) {
	ctx, span := GetSpanFromContext(ctx, db.cfg, sql, "EXECUTE", 2)
	defer span.End()

	dbParam, err := implement.ResolveParams(input, db.tpl.StmtDbTypeWrap)
	if err != nil {
		return
	}
	query, execArgs, err := db.tpl.GetSQLContext(sql, dbParam, opts...)
	if err != nil {
		err = implement.GetError(err, sql, input)
		return
	}

	start := time.Now()
	debugPrint(ctx, db.cfg, query, execArgs...)
	r, err = db.db.Exec(ctx, query, execArgs...)
	if err != nil {
		return r, implement.GetError(err, query, execArgs...)
	}
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)

	return
}

// Query 查询数据
func (db *xDB) QueryAs(ctx context.Context, sqls string, input any, results any, opts ...xdb.TemplateOption) (err error) {
	return db.dbQueryAs(ctx, sqls, input, results, func(r *sql.Rows, val any) error {
		return implement.ResolveRowsDataResult(db.proto, r, val)
	}, opts...)
}

func (db *xDB) FirstAs(ctx context.Context, sqls string, input any, result any, opts ...xdb.TemplateOption) (err error) {
	return db.dbQueryAs(ctx, sqls, input, result, func(r *sql.Rows, val any) error {
		return implement.ResolveFirstDataResult(db.proto, r, val)
	}, opts...)
}

// Begin 创建事务
func (db *xDB) Begin() (t xdb.ITrans, err error) {
	return db.BeginTx(context.Background())
}

func (db *xDB) BeginTx(ctx context.Context) (t xdb.ITrans, err error) {
	ctx, span := GetSpanFromContext(ctx, db.cfg, "", "BeginTx", 2)
	defer span.End()
	return db.createTrans(ctx)
}

// Transaction 执行事务
func (db *xDB) Transaction(ctx context.Context, callback xdb.TransactionCallback) (err error) {
	ctx, span := GetSpanFromContext(ctx, db.cfg, "", "Transaction", 2)
	defer span.End()

	tx, err := db.createTrans(ctx)
	if err != nil {
		return
	}
	defer func() {
		if robj := recover(); robj != nil {
			tx.Rollback()
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
	err = callback(ctx, tx)
	if err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	return
}

// Close  关闭当前数据库连接
func (db *xDB) Close() error {
	return db.db.Close()
}

func (db *xDB) dbQuery(ctx context.Context, sql string, input any, callback implement.DbResolveMapValCallback, opts ...xdb.TemplateOption) (result any, err error) {
	ctx, span := GetSpanFromContext(ctx, db.cfg, sql, "SELECT", 3)
	defer span.End()

	dbParams, err := implement.ResolveParams(input, db.tpl.StmtDbTypeWrap)
	if err != nil {
		return
	}

	query, execArgs, err := db.tpl.GetSQLContext(sql, dbParams, opts...)
	if err != nil {
		err = implement.GetError(err, sql, input)
		return
	}

	start := time.Now()

	debugPrint(ctx, db.cfg, query, execArgs...)
	rows, err := db.db.Query(ctx, query, execArgs...)
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

func (db *xDB) dbQueryAs(ctx context.Context, sql string, input any, result any, callback implement.DbResolveResultCallback, opts ...xdb.TemplateOption) (err error) {
	ctx, span := GetSpanFromContext(ctx, db.cfg, sql, "SELECT", 3)
	defer span.End()

	dbParams, err := implement.ResolveParams(input, db.tpl.StmtDbTypeWrap)
	if err != nil {
		return
	}

	query, execArgs, err := db.tpl.GetSQLContext(sql, dbParams, opts...)
	if err != nil {
		err = implement.GetError(err, sql, input)
		return
	}

	start := time.Now()

	debugPrint(ctx, db.cfg, query, execArgs...)
	rows, err := db.db.Query(ctx, query, execArgs...)
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

func (db *xDB) createTrans(ctx context.Context) (t xdb.ITrans, err error) {
	tt := &xTrans{
		cfg:   db.cfg,
		proto: db.proto,
	}
	tt.tx, err = db.db.BeginTx(ctx)
	if err != nil {
		return
	}
	tt.tpl = db.tpl
	return tt, nil
}
