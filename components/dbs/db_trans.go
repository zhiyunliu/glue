package dbs

import (
	"github.com/zhiyunliu/gel/components/dbs/tpl"
)

//DBTrans 数据库事务操作类
type DBTrans struct {
	tpl tpl.ITPLContext
	tx  ISysTrans
}

//Query 查询数据
func (t *DBTrans) Query(sql string, input map[string]interface{}) (data Rows, err error) {
	query, args := t.tpl.GetSQLContext(sql, input)
	data, err = t.tx.Query(query, args...)
	if err != nil {
		return nil, getDBError(err, query, args)
	}
	return
}

//Scalar 根据包含@名称占位符的查询语句执行查询语句
func (t *DBTrans) Scalar(sql string, input map[string]interface{}) (data interface{}, err error) {
	query, args := t.tpl.GetSQLContext(sql, input)
	result, err := t.tx.Query(query, args...)
	if err != nil {
		return nil, getDBError(err, query, args)
	}
	if result.Len() == 0 || result[0].Len() == 0 || len(result[0].Keys()) == 0 {
		return nil, nil
	}
	data, _ = result[0].Get(result[0].Keys()[0])
	return
}

//Execute 根据包含@名称占位符的语句执行查询语句
func (t *DBTrans) Execute(sql string, input map[string]interface{}) (r Result, err error) {
	query, args := t.tpl.GetSQLContext(sql, input)
	r, err = t.tx.Execute(query, args...)
	if err != nil {
		return nil, getDBError(err, query, args)
	}
	return
}

//ExecuteSP 根据包含@名称占位符的语句执行查询语句
func (t *DBTrans) ExecuteSP(sql string, input map[string]interface{}) (r Result, err error) {
	query, args := t.tpl.GetSPContext(sql, input)
	r, err = t.tx.Execute(query, args...)
	if err != nil {
		return nil, getDBError(err, query, args)
	}
	return
}

//Rollback 回滚所有操作
func (t *DBTrans) Rollback() error {
	return t.tx.Rollback()
}

//Commit 提交所有操作
func (t *DBTrans) Commit() error {
	return t.tx.Commit()
}
