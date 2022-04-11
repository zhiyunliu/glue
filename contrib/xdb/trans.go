package xdb

import (
	"github.com/zhiyunliu/gel/contrib/xdb/internal"
	"github.com/zhiyunliu/gel/contrib/xdb/tpl"
	"github.com/zhiyunliu/gel/xdb"
	"github.com/zhiyunliu/golibs/xtypes"
)

//xTrans 数据库事务操作类
type xTrans struct {
	tpl tpl.SQLTemplate
	tx  internal.ISysTrans
}

//Query 查询数据
func (t *xTrans) Query(sql string, input map[string]interface{}) (rows xdb.Rows, err error) {
	query, args := t.tpl.GetSQLContext(sql, input)
	data, err := t.tx.Query(query, args...)
	if err != nil {
		return nil, getError(err, query, args)
	}

	rows, err = resolveRows(data, 0)
	if err != nil {
		return nil, getError(err, query, args)
	}
	return
}

func (t *xTrans) First(sql string, input map[string]interface{}) (data xdb.Row, err error) {
	rows, err := t.Query(sql, input)
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

//Scalar 根据包含@名称占位符的查询语句执行查询语句
func (t *xTrans) Scalar(sql string, input map[string]interface{}) (result interface{}, err error) {
	rows, err := t.Query(sql, input)
	if err != nil {
		return
	}
	if rows.Len() == 0 || len(rows[0]) == 0 {
		return nil, nil
	}
	result, _ = rows[0].Get(rows[0].Keys()[0])
	return
}

//Execute 根据包含@名称占位符的语句执行查询语句
func (t *xTrans) Exec(sql string, input map[string]interface{}) (r xdb.Result, err error) {
	query, args := t.tpl.GetSQLContext(sql, input)
	r, err = t.tx.Execute(query, args...)
	if err != nil {
		return nil, getError(err, query, args)
	}
	return
}

//Rollback 回滚所有操作
func (t *xTrans) Rollback() error {
	return t.tx.Rollback()
}

//Commit 提交所有操作
func (t *xTrans) Commit() error {
	return t.tx.Commit()
}
