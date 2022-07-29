package xgorm

import (
	"context"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xtypes"
	"gorm.io/gorm"
)

var _ xdb.IDB = &dbWrap{}

type dbWrap struct {
	gromDB *gorm.DB
}

func (d *dbWrap) Begin() (xdb.ITrans, error) {
	txdb := d.gromDB.Begin()
	return &transWrap{
		gromDB: txdb,
	}, nil
}

func (d *dbWrap) Close() (err error) {
	return
}
func (d *dbWrap) GetImpl() interface{} {
	return d.gromDB
}
func (d *dbWrap) Query(ctx context.Context, sql string, input map[string]interface{}) (data xdb.Rows, err error) {
	data = make(xtypes.XMaps, 0)
	err = d.gromDB.Raw(sql, input).Scan(&data).Error
	return

}
func (d *dbWrap) Multi(ctx context.Context, sql string, input map[string]interface{}) (data []xdb.Rows, err error) {
	rows, err := d.gromDB.Raw(sql, input).Rows()
	if err != nil {
		err = internal.GetError(err, sql, []interface{}{input})
		return
	}
	data, err = internal.ResolveMultiRows(rows)
	return
}
func (d *dbWrap) First(ctx context.Context, sql string, input map[string]interface{}) (data xdb.Row, err error) {
	data = make(xtypes.XMap)
	err = d.gromDB.Raw(sql, input).First(&data).Error
	if err != nil {
		err = internal.GetError(err, sql, []interface{}{input})
		return
	}
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
	tmpDb := d.gromDB.Exec(sql, input)
	err = tmpDb.Error
	if err != nil {
		return
	}
	return &sqlResult{
		rowsAffected: tmpDb.RowsAffected,
	}, err
}
