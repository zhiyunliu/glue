package xgorm

import (
	"context"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
	"gorm.io/gorm"
)

var _ xdb.ITrans = &transWrap{}

type transWrap struct {
	gromDB *gorm.DB
	tpl    tpl.SQLTemplate
}

func (d *transWrap) Rollback() (err error) {
	err = d.gromDB.Rollback().Error
	return
}
func (d *transWrap) Commit() (err error) {
	err = d.gromDB.Commit().Error
	return
}

func (d *transWrap) Query(ctx context.Context, sql string, input map[string]interface{}) (data xdb.Rows, err error) {
	query, args := d.tpl.GetSQLContext(sql, input)
	rows, err := d.gromDB.Raw(query, args...).Rows()
	if err != nil {
		err = internal.GetError(err, query, args...)
		return
	}
	data, err = internal.ResolveRows(rows)
	return

}
func (d *transWrap) Multi(ctx context.Context, sql string, input map[string]interface{}) (data []xdb.Rows, err error) {
	query, args := d.tpl.GetSQLContext(sql, input)
	rows, err := d.gromDB.Raw(query, args...).Rows()
	if err != nil {
		err = internal.GetError(err, query, args...)
		return
	}
	data, err = internal.ResolveMultiRows(rows)
	return
}
func (d *transWrap) First(ctx context.Context, sql string, input map[string]interface{}) (data xdb.Row, err error) {
	rows, err := d.Query(ctx, sql, input)
	if err != nil {
		return
	}
	data = rows.Get(0)
	return
}
func (d *transWrap) Scalar(ctx context.Context, sql string, input map[string]interface{}) (data interface{}, err error) {

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
func (d *transWrap) Exec(ctx context.Context, sql string, input map[string]interface{}) (r xdb.Result, err error) {
	query, args := d.tpl.GetSQLContext(sql, input)

	tmpDb := d.gromDB.Exec(query, args...)
	err = tmpDb.Error
	if err != nil {
		err = internal.GetError(err, query, args...)
		return
	}
	return &sqlResult{
		rowsAffected: tmpDb.RowsAffected,
	}, err
}
