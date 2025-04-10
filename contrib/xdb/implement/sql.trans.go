package implement

import (
	"context"
	"database/sql"
)

// ISysTrans 数据库事务接口
type ISysTrans interface {
	Query(context.Context, string, ...interface{}) (*sql.Rows, error)
	Execute(context.Context, string, ...interface{}) (sql.Result, error)
	Rollback() error
	Commit() error
}

// SysTrans 事务
type sysTrans struct {
	tx *sql.Tx
}

// Query 执行查询
func (t *sysTrans) Query(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	rows, err = t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return
	}
	return
}

// Executes 执行SQL操作语句
func (t *sysTrans) Execute(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error) {
	result, err = t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return
	}
	return result, nil
}

// Rollback 回滚所有操作
func (t *sysTrans) Rollback() error {
	return t.tx.Rollback()
}

// Commit 提交所有操作
func (t *sysTrans) Commit() error {
	return t.tx.Commit()
}
