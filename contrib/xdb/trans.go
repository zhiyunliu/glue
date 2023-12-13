package xdb

import (
	"context"
	"database/sql"
	"time"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

// xTrans 数据库事务操作类
type xTrans struct {
	cfg *Setting
	tpl tpl.SQLTemplate
	tx  internal.ISysTrans
}

// Query 查询数据
func (db *xTrans) Query(ctx context.Context, sqls string, input any) (rows xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return internal.ResolveRows(r)
	})
	rows = tmp.(xdb.Rows)
	return
}

// Query 查询数据
func (db *xTrans) Multi(ctx context.Context, sqls string, input any) (datasetRows []xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return internal.ResolveMultiRows(r)
	})
	datasetRows = tmp.([]xdb.Rows)
	return
}

func (t *xTrans) First(ctx context.Context, sqls string, input any) (data xdb.Row, err error) {
	tmp, err := t.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return internal.ResolveFirstRow(r)
	})
	data = tmp.(xdb.Row)
	return
}

// Scalar 根据包含@名称占位符的查询语句执行查询语句
func (t *xTrans) Scalar(ctx context.Context, sqls string, input any) (data interface{}, err error) {
	data, err = t.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return internal.ResolveScalar(r)
	})
	return
}

// Execute 根据包含@名称占位符的语句执行查询语句
func (db *xTrans) Exec(ctx context.Context, sql string, input any) (r xdb.Result, err error) {

	dbParams, err := internal.ResolveParams(input)
	if err != nil {
		return
	}
	query, execArgs := db.tpl.GetSQLContext(sql, dbParams)
	start := time.Now()
	debugPrint(ctx, db.cfg, query, execArgs...)
	r, err = db.tx.Execute(query, execArgs...)
	if err != nil {
		return nil, internal.GetError(err, query, execArgs...)
	}
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)
	return
}

// Query 查询数据
func (db *xTrans) QueryAs(ctx context.Context, sqls string, input any, results any) (err error) {
	return db.dbQueryAs(ctx, sqls, input, results, func(r *sql.Rows, a any) error {
		return internal.ResolveRowsDataResult(r, results)
	})
}

func (db *xTrans) FirstAs(ctx context.Context, sqls string, input any, result any) (err error) {
	return db.dbQueryAs(ctx, sqls, input, result, func(r *sql.Rows, a any) error {
		return internal.ResolveFirstDataResult(r, result)
	})
}

// Rollback 回滚所有操作
func (t *xTrans) Rollback() error {
	return t.tx.Rollback()
}

// Commit 提交所有操作
func (t *xTrans) Commit() error {
	return t.tx.Commit()
}

func (db *xTrans) dbQuery(ctx context.Context, sql string, input any, callback internal.DbResolveMapValCallback) (result any, err error) {
	dbParams, err := internal.ResolveParams(input)
	if err != nil {
		return
	}

	query, execArgs := db.tpl.GetSQLContext(sql, dbParams)

	start := time.Now()

	debugPrint(ctx, db.cfg, query, execArgs...)
	rows, err := db.tx.Query(query, execArgs...)
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

func (db *xTrans) dbQueryAs(ctx context.Context, sql string, input any, result any, callback internal.DbResolveResultCallback) (err error) {
	dbParams, err := internal.ResolveParams(input)
	if err != nil {
		return
	}

	query, execArgs := db.tpl.GetSQLContext(sql, dbParams)

	start := time.Now()

	debugPrint(ctx, db.cfg, query, execArgs...)
	rows, err := db.tx.Query(query, execArgs...)
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
