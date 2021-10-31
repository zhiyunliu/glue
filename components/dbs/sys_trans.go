package dbs

import "database/sql"

//ISysDBTrans 数据库事务接口
type ISysTrans interface {
	Query(string, ...interface{}) (Rows, error)
	Execute(string, ...interface{}) (Result, error)
	Rollback() error
	Commit() error
}

//SysTrans 事务
type SysTrans struct {
	tx *sql.Tx
}

//Query 执行查询
func (t *SysTrans) Query(query string, args ...interface{}) (dataRows Rows, err error) {
	rows, err := t.tx.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()
	dataRows, _, err = resolveRows(rows, 0)
	return
}

//Executes 执行SQL操作语句
func (t *SysTrans) Execute(query string, args ...interface{}) (r Result, err error) {
	result, err := t.tx.Exec(query, args...)
	if err != nil {
		return
	}
	return result, nil
}

//Rollback 回滚所有操作
func (t *SysTrans) Rollback() error {
	return t.tx.Rollback()
}

//Commit 提交所有操作
func (t *SysTrans) Commit() error {
	return t.tx.Commit()
}
