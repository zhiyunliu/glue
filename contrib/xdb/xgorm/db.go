package xgorm

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"

	"github.com/zhiyunliu/glue/contrib/xdb/implement"
	"github.com/zhiyunliu/glue/xdb"
	"gorm.io/gorm"
)

var _ xdb.IDB = &dbWrap{}

type dbWrap struct {
	gromDB *gorm.DB
	proto  string
	tpl    xdb.SQLTemplate
}

func (d *dbWrap) Begin() (xdb.ITrans, error) {
	return d.BeginTx(context.Background())
}

func (d *dbWrap) BeginTx(ctx context.Context) (xdb.ITrans, error) {
	txdb := d.gromDB.Begin()
	return &transWrap{
		gromDB: txdb,
		proto:  d.proto,
		tpl:    d.tpl,
	}, nil
}

func (d *dbWrap) Close() (err error) {
	return
}

func (d *dbWrap) GetImpl() interface{} {
	return d.gromDB
}

func (db *dbWrap) Query(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (data xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveRows(db.proto, r)
	}, opts...)
	if err != nil {
		return
	}
	data = tmp.(xdb.Rows)
	return
}

func (db *dbWrap) Multi(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (data []xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveMultiRows(db.proto, r)
	}, opts...)
	if err != nil {
		return
	}
	data = tmp.([]xdb.Rows)
	return
}

func (db *dbWrap) First(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (data xdb.Row, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveFirstRow(db.proto, r)
	}, opts...)
	if err != nil {
		return
	}
	data = tmp.(xdb.Row)
	return
}

func (db *dbWrap) Scalar(ctx context.Context, sqls string, input any, opts ...xdb.TemplateOption) (data interface{}, err error) {
	data, err = db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveScalar(db.proto, r)
	}, opts...)
	return
}

func (d *dbWrap) Exec(ctx context.Context, sql string, input any, opts ...xdb.TemplateOption) (r xdb.Result, err error) {
	dbParam, err := implement.ResolveParams(input, d.tpl.StmtDbTypeWrap)
	if err != nil {
		return
	}
	query, execArgs, err := d.tpl.GetSQLContext(sql, dbParam, opts...)
	if err != nil {
		err = implement.GetError(err, sql, input)
		return
	}

	db := d.gromDB
	db = db.Exec(query, execArgs...)
	err = db.Error
	if err != nil {
		err = implement.GetError(err, query, execArgs...)
		return
	}

	r = &sqlResult{
		rowsAffected: db.RowsAffected,
	}
	return
}

// Query 查询数据
func (db *dbWrap) QueryAs(ctx context.Context, sqls string, input any, results any, opts ...xdb.TemplateOption) (err error) {
	return db.dbQueryAs(ctx, sqls, input, results, func(r *sql.Rows, a any) error {
		return implement.ResolveRowsDataResult(db.proto, r, results)
	}, opts...)
}

func (db *dbWrap) FirstAs(ctx context.Context, sqls string, input any, result any, opts ...xdb.TemplateOption) (err error) {
	return db.dbQueryAs(ctx, sqls, input, result, func(r *sql.Rows, a any) error {
		return implement.ResolveFirstDataResult(db.proto, r, result)
	}, opts...)
}

// Transaction 执行事务
func (d *dbWrap) Transaction(ctx context.Context, callback xdb.TransactionCallback) (err error) {
	txdb := d.gromDB.Begin()
	tt := &transWrap{
		gromDB: txdb,
		tpl:    d.tpl,
	}

	defer func() {
		if robj := recover(); robj != nil {
			tt.Rollback()
			rerr, ok := robj.(error)
			if !ok {
				rerr = fmt.Errorf("%+v", robj)
			}
			buf := make([]byte, 64<<10) //nolint:gomnd
			n := runtime.Stack(buf, false)
			buf = buf[:n]
			err = xdb.NewPanicError(rerr, string(buf))
		}
	}()
	err = callback(ctx, tt)
	if err != nil {
		tt.Rollback()
		return
	}
	err = tt.Commit()
	return
}

func (db *dbWrap) dbQuery(ctx context.Context, sql string, input any, callback implement.DbResolveMapValCallback, opts ...xdb.TemplateOption) (result any, err error) {
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

func (db *dbWrap) dbQueryAs(ctx context.Context, sql string, input any, result any, callback implement.DbResolveResultCallback, opts ...xdb.TemplateOption) (err error) {
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
