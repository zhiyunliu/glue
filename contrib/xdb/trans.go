package xdb

import (
	"context"
	"time"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xtypes"
)

// xTrans 数据库事务操作类
type xTrans struct {
	cfg *Setting
	tpl tpl.SQLTemplate
	tx  internal.ISysTrans
}

// Query 查询数据
func (db *xTrans) Query(ctx context.Context, sql string, input map[string]interface{}) (rows xdb.Rows, err error) {
	start := time.Now()
	query, execArgs := db.tpl.GetSQLContext(sql, input)

	debugPrint(ctx, db.cfg, query, execArgs...)
	data, err := db.tx.Query(query, execArgs...)
	if err != nil {
		return nil, internal.GetError(err, query, execArgs...)
	}
	defer func() {
		if data != nil {
			data.Close()
		}
	}()
	rows, err = internal.ResolveRows(data)
	if err != nil {
		return nil, internal.GetError(err, query, execArgs...)
	}
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)
	return
}

// Query 查询数据
func (db *xTrans) Multi(ctx context.Context, sql string, input map[string]interface{}) (datasetRows []xdb.Rows, err error) {
	start := time.Now()

	query, execArgs := db.tpl.GetSQLContext(sql, input)

	debugPrint(ctx, db.cfg, query, execArgs...)

	sqlRows, err := db.tx.Query(query, execArgs...)
	if err != nil {
		return nil, internal.GetError(err, query, execArgs...)
	}
	defer func() {
		if sqlRows != nil {
			sqlRows.Close()
		}
	}()
	datasetRows, err = internal.ResolveMultiRows(sqlRows)
	if err != nil {
		return nil, internal.GetError(err, query, execArgs...)
	}
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)
	return
}

func (t *xTrans) First(ctx context.Context, sql string, input map[string]interface{}) (data xdb.Row, err error) {
	rows, err := t.Query(ctx, sql, input)
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

// Scalar 根据包含@名称占位符的查询语句执行查询语句
func (t *xTrans) Scalar(ctx context.Context, sql string, input map[string]interface{}) (result interface{}, err error) {
	rows, err := t.Query(ctx, sql, input)
	if err != nil {
		return
	}
	if rows.Len() == 0 || len(rows[0]) == 0 {
		return nil, nil
	}
	result, _ = rows[0].Get(rows[0].Keys()[0])
	return
}

// Execute 根据包含@名称占位符的语句执行查询语句
func (db *xTrans) Exec(ctx context.Context, sql string, input map[string]interface{}) (r xdb.Result, err error) {
	start := time.Now()
	query, execArgs := db.tpl.GetSQLContext(sql, input)

	debugPrint(ctx, db.cfg, query, execArgs...)
	r, err = db.tx.Execute(query, execArgs...)
	if err != nil {
		return nil, internal.GetError(err, query, execArgs...)
	}
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)
	return
}

// Rollback 回滚所有操作
func (t *xTrans) Rollback() error {
	return t.tx.Rollback()
}

// Commit 提交所有操作
func (t *xTrans) Commit() error {
	return t.tx.Commit()
}
