package xdb

import (
	"context"
	"database/sql"
	"time"

	"github.com/zhiyunliu/glue/contrib/xdb/implement"
	"github.com/zhiyunliu/glue/xdb"
)

// xTrans 数据库事务操作类
type xTrans struct {
	cfg   *Setting
	proto string
	tpl   xdb.SQLTemplate
	tx    implement.ISysTrans
}

// Query 查询数据
func (db *xTrans) Query(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (rows xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveRows(db.proto, r)
	}, opts...)
	if err != nil {
		return
	}
	rows = tmp.(xdb.Rows)
	return
}

// Query 查询数据
func (db *xTrans) Multi(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (datasetRows []xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveMultiRows(db.proto, r)
	}, opts...)
	if err != nil {
		return
	}
	datasetRows = tmp.([]xdb.Rows)
	return
}

func (db *xTrans) First(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (data xdb.Row, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveFirstRow(db.proto, r)
	}, opts...)
	if err != nil {
		return
	}
	data = tmp.(xdb.Row)
	return
}

// Scalar 根据包含@名称占位符的查询语句执行查询语句
func (db *xTrans) Scalar(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (data interface{}, err error) {
	data, err = db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveScalar(db.proto, r)
	}, opts...)
	return
}

// Execute 根据包含@名称占位符的语句执行查询语句
func (db *xTrans) Exec(ctx context.Context, sql string, input any, opts ...xdb.TemplateOption) (r xdb.Result, err error) {

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
	r, err = db.tx.Execute(ctx, query, execArgs...)
	if err != nil {
		return nil, implement.GetError(err, query, execArgs...)
	}
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)
	return
}

// Query 查询数据
func (db *xTrans) QueryAs(ctx context.Context, sqls string, input any, results any, opts ...xdb.TemplateOption) (err error) {
	return db.dbQueryAs(ctx, sqls, input, results, func(r *sql.Rows, a any) error {
		return implement.ResolveRowsDataResult(db.proto, r, results)
	}, opts...)
}

func (db *xTrans) FirstAs(ctx context.Context, sqls string, input any, result any, opts ...xdb.TemplateOption) (err error) {
	return db.dbQueryAs(ctx, sqls, input, result, func(r *sql.Rows, a any) error {
		return implement.ResolveFirstDataResult(db.proto, r, result)
	}, opts...)
}

// Rollback 回滚所有操作
func (t *xTrans) Rollback() error {
	return t.tx.Rollback()
}

// Commit 提交所有操作
func (t *xTrans) Commit() error {
	return t.tx.Commit()
}

func (db *xTrans) dbQuery(ctx context.Context, sql string, input any, callback implement.DbResolveMapValCallback, opts ...xdb.TemplateOption) (result any, err error) {
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
	rows, err := db.tx.Query(ctx, query, execArgs...)
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

func (db *xTrans) dbQueryAs(ctx context.Context, sql string, input any, result any, callback implement.DbResolveResultCallback, opts ...xdb.TemplateOption) (err error) {
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
	rows, err := db.tx.Query(ctx, query, execArgs...)
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
