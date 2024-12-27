package xgorm

import (
	"context"
	"database/sql"

	"github.com/zhiyunliu/glue/contrib/xdb/implement"
	"github.com/zhiyunliu/glue/xdb"
	"gorm.io/gorm"
)

var _ xdb.ITrans = &transWrap{}

type transWrap struct {
	gromDB *gorm.DB
	proto  string
	tpl    xdb.SQLTemplate
}

func (d *transWrap) Rollback() (err error) {
	err = d.gromDB.Rollback().Error
	return
}
func (d *transWrap) Commit() (err error) {
	err = d.gromDB.Commit().Error
	return
}

func (db *transWrap) Query(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (data xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveRows(db.proto, r)
	}, opts...)
	if err != nil {
		return
	}
	data = tmp.(xdb.Rows)
	return
}

func (db *transWrap) Multi(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (data []xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveMultiRows(db.proto, r)
	}, opts...)
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

func (db *transWrap) First(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (data xdb.Row, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveFirstRow(db.proto, r)
	}, opts...)
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

func (db *transWrap) Scalar(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (data interface{}, err error) {
	data, err = db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveScalar(db.proto, r)
	}, opts...)
	return
}

func (d *transWrap) Exec(ctx context.Context, sql string, input any, opts ...xdb.TemplateOption) (r xdb.Result, err error) {

	dbParam, err := implement.ResolveParams(input, d.tpl.StmtDbTypeWrap)
	if err != nil {
		return
	}
	query, execArgs, err := d.tpl.GetSQLContext(sql, dbParam, opts...)
	if err != nil {
		err = implement.GetError(err, sql, input)
		return
	}

	tmpDb := d.gromDB.Exec(query, execArgs...)
	err = tmpDb.Error
	if err != nil {
		err = implement.GetError(err, query, execArgs...)
		return
	}
	return &sqlResult{
		rowsAffected: tmpDb.RowsAffected,
	}, err
}

// Query 查询数据
func (db *transWrap) QueryAs(ctx context.Context, sqls string, input any, results any, opts ...xdb.TemplateOption) (err error) {
	return db.dbQueryAs(ctx, sqls, input, results, func(r *sql.Rows, a any) error {
		return implement.ResolveRowsDataResult(db.proto, r, results)
	}, opts...)
}

func (db *transWrap) FirstAs(ctx context.Context, sqls string, input any, result any, opts ...xdb.TemplateOption) (err error) {
	return db.dbQueryAs(ctx, sqls, input, result, func(r *sql.Rows, a any) error {
		return implement.ResolveFirstDataResult(db.proto, r, result)
	}, opts...)
}

func (db *transWrap) dbQuery(ctx context.Context, sql string, input any, callback implement.DbResolveMapValCallback, opts ...xdb.TemplateOption) (result any, err error) {
	dbParam, err := implement.ResolveParams(input, db.tpl.StmtDbTypeWrap)
	if err != nil {
		return
	}

	query, execArgs, err := db.tpl.GetSQLContext(sql, dbParam, opts...)
	if err != nil {
		err = implement.GetError(err, sql, input)
		return
	}

	rows, err := db.gromDB.Raw(query, execArgs...).Rows()
	if err != nil {
		err = implement.GetError(err, query, execArgs...)
		return
	}
	defer rows.Close()
	result, err = callback(rows)
	return
}

func (db *transWrap) dbQueryAs(_ context.Context, sql string, input any, result any, callback implement.DbResolveResultCallback, opts ...xdb.TemplateOption) (err error) {
	dbParam, err := implement.ResolveParams(input, db.tpl.StmtDbTypeWrap)
	if err != nil {
		return
	}

	query, execArgs, err := db.tpl.GetSQLContext(sql, dbParam, opts...)
	if err != nil {
		err = implement.GetError(err, sql, input)
		return
	}

	rows, err := db.gromDB.Raw(query, execArgs...).Rows()
	if err != nil {
		err = implement.GetError(err, query, execArgs...)
		return
	}
	defer rows.Close()
	err = callback(rows, result)
	return
}
