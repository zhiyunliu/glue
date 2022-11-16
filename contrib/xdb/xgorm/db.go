package xgorm

import (
	"context"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
	"gorm.io/gorm"
)

var _ xdb.IDB = &dbWrap{}

type dbWrap struct {
	gromDB *gorm.DB
	tpl    tpl.SQLTemplate
}

func (d *dbWrap) Begin() (xdb.ITrans, error) {
	txdb := d.gromDB.Begin()
	return &transWrap{
		gromDB: txdb,
		tpl:    d.tpl,
	}, nil
}

func (d *dbWrap) Close() (err error) {
	return
}
func (d *dbWrap) GetImpl() interface{} {
	return d.gromDB
}
func (d *dbWrap) Query(ctx context.Context, sql string, input map[string]interface{}) (data xdb.Rows, err error) {
	query, args := d.tpl.GetSQLContext(sql, input)
	rows, err := d.gromDB.Raw(query, args...).Rows()
	if err != nil {
		err = internal.GetError(err, query, args...)
		return
	}
	data, err = internal.ResolveRows(rows)
	return

}
func (d *dbWrap) Multi(ctx context.Context, sql string, input map[string]interface{}) (data []xdb.Rows, err error) {
	query, args := d.tpl.GetSQLContext(sql, input)
	rows, err := d.gromDB.Raw(query, args...).Rows()
	if err != nil {
		err = internal.GetError(err, query, args...)
		return
	}
	data, err = internal.ResolveMultiRows(rows)
	return
}
func (d *dbWrap) First(ctx context.Context, sql string, input map[string]interface{}) (data xdb.Row, err error) {
	rows, err := d.Query(ctx, sql, input)
	if err != nil {
		return
	}
	data = rows.Get(0)
	return
}
func (d *dbWrap) Scalar(ctx context.Context, sql string, input map[string]interface{}) (data interface{}, err error) {

	row, err := d.First(ctx, sql, input)
	if err != nil {
		return
	}
	if row.Len() == 0 {
		return nil, nil
	}
	data, _ = row.Get(row.Keys()[0])
	return
}
func (d *dbWrap) Exec(ctx context.Context, sql string, input map[string]interface{}) (r xdb.Result, err error) {
	query, args := d.tpl.GetSQLContext(sql, input)

	db := d.gromDB

	result, err := db.Statement.ConnPool.ExecContext(db.Statement.Context, query, args...)
	if err != nil {
		err = internal.GetError(err, query, args...)
		return
	}
	r = result
	return
}
