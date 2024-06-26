package xgorm

import (
	"context"
	"database/sql"

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

func (db *transWrap) Query(ctx context.Context, sqls string, input any) (data xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return internal.ResolveRows(r)
	})
	if err != nil {
		return
	}
	data = tmp.(xdb.Rows)
	return
}

func (db *transWrap) Multi(ctx context.Context, sqls string, input any) (data []xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return internal.ResolveMultiRows(r)
	})
	if err != nil {
		return
	}
	data = tmp.([]xdb.Rows)
	return
}

// func (d *transWrap) First(ctx context.Context, sql string, input any) (data xdb.Row, err error) {
// 	rows, err := d.Query(ctx, sql, input)
// 	if err != nil {
// 		return
// 	}
// 	data = rows.Get(0)
// 	return
// }

func (db *transWrap) First(ctx context.Context, sqls string, input any) (data xdb.Row, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return internal.ResolveFirstRow(r)
	})
	if err != nil {
		return
	}
	data = tmp.(xdb.Row)
	return
}

// func (d *transWrap) Scalar(ctx context.Context, sql string, input any) (data interface{}, err error) {

// 	row, err := d.First(ctx, sql, input)
// 	if err != nil {
// 		return
// 	}
// 	if row.Len() == 0 {
// 		return nil, nil
// 	}
// 	data, _ = row.Get(row.Keys()[0])
// 	return
// }

func (db *transWrap) Scalar(ctx context.Context, sqls string, input any) (data interface{}, err error) {
	data, err = db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return internal.ResolveScalar(r)
	})
	return
}

func (d *transWrap) Exec(ctx context.Context, sql string, input any) (r xdb.Result, err error) {

	dbParam, err := internal.ResolveParams(input)
	if err != nil {
		return
	}
	query, execArgs, err := d.tpl.GetSQLContext(sql, dbParam)
	if err != nil {
		err = internal.GetError(err, sql, input)
		return
	}

	tmpDb := d.gromDB.Exec(query, execArgs...)
	err = tmpDb.Error
	if err != nil {
		err = internal.GetError(err, query, execArgs...)
		return
	}
	return &sqlResult{
		rowsAffected: tmpDb.RowsAffected,
	}, err
}

// Query 查询数据
func (d *transWrap) QueryAs(ctx context.Context, sqls string, input any, results any) (err error) {
	return d.dbQueryAs(ctx, sqls, input, results, func(r *sql.Rows, a any) error {
		return internal.ResolveRowsDataResult(r, results)
	})
}

func (d *transWrap) FirstAs(ctx context.Context, sqls string, input any, result any) (err error) {
	return d.dbQueryAs(ctx, sqls, input, result, func(r *sql.Rows, a any) error {
		return internal.ResolveFirstDataResult(r, result)
	})
}

func (db *transWrap) dbQuery(ctx context.Context, sql string, input any, callback internal.DbResolveMapValCallback) (result any, err error) {
	dbParam, err := internal.ResolveParams(input)
	if err != nil {
		return
	}

	query, execArgs, err := db.tpl.GetSQLContext(sql, dbParam)
	if err != nil {
		err = internal.GetError(err, sql, input)
		return
	}

	rows, err := db.gromDB.Raw(query, execArgs...).Rows()
	if err != nil {
		err = internal.GetError(err, query, execArgs...)
		return
	}
	result, err = callback(rows)
	return
}

func (db *transWrap) dbQueryAs(ctx context.Context, sql string, input any, result any, callback internal.DbResolveResultCallback) (err error) {
	dbParam, err := internal.ResolveParams(input)
	if err != nil {
		return
	}

	query, execArgs, err := db.tpl.GetSQLContext(sql, dbParam)
	if err != nil {
		err = internal.GetError(err, sql, input)
		return
	}

	rows, err := db.gromDB.Raw(query, execArgs...).Rows()
	if err != nil {
		err = internal.GetError(err, query, execArgs...)
		return
	}
	err = callback(rows, result)
	return
}
