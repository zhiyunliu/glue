package xgorm

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"

	"github.com/zhiyunliu/glue/contrib/xdb/implement"
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

func (db *dbWrap) Query(ctx context.Context, sqls string, input any) (data xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveRows(r)
	})
	if err != nil {
		return
	}
	data = tmp.(xdb.Rows)
	return
}

func (db *dbWrap) Multi(ctx context.Context, sqls string, input any) (data []xdb.Rows, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveMultiRows(r)
	})
	if err != nil {
		return
	}
	data = tmp.([]xdb.Rows)
	return
}

func (db *dbWrap) First(ctx context.Context, sqls string, input any) (data xdb.Row, err error) {
	tmp, err := db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveFirstRow(r)
	})
	if err != nil {
		return
	}
	data = tmp.(xdb.Row)
	return
}

func (db *dbWrap) Scalar(ctx context.Context, sqls string, input any) (data interface{}, err error) {
	data, err = db.dbQuery(ctx, sqls, input, func(r *sql.Rows) (any, error) {
		return implement.ResolveScalar(r)
	})
	return
}

func (d *dbWrap) Exec(ctx context.Context, sql string, input any) (r xdb.Result, err error) {
	dbParam, err := implement.ResolveParams(input)
	if err != nil {
		return
	}
	query, execArgs, err := d.tpl.GetSQLContext(sql, dbParam)
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
func (db *dbWrap) QueryAs(ctx context.Context, sqls string, input any, results any) (err error) {
	return db.dbQueryAs(ctx, sqls, input, results, func(r *sql.Rows, a any) error {
		return implement.ResolveRowsDataResult(r, results)
	})
}

func (db *dbWrap) FirstAs(ctx context.Context, sqls string, input any, result any) (err error) {
	return db.dbQueryAs(ctx, sqls, input, result, func(r *sql.Rows, a any) error {
		return implement.ResolveFirstDataResult(r, result)
	})
}

// Transaction 执行事务
func (d *dbWrap) Transaction(callback xdb.TransactionCallback) (err error) {
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
	err = callback(tt)
	if err != nil {
		tt.Rollback()
		return
	}
	tt.Commit()
	return
}

func (db *dbWrap) dbQuery(ctx context.Context, sql string, input any, callback implement.DbResolveMapValCallback) (result any, err error) {
	dbParam, err := implement.ResolveParams(input)
	if err != nil {
		return
	}

	query, execArgs, err := db.tpl.GetSQLContext(sql, dbParam)
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

func (db *dbWrap) dbQueryAs(ctx context.Context, sql string, input any, result any, callback implement.DbResolveResultCallback) (err error) {
	dbParam, err := implement.ResolveParams(input)
	if err != nil {
		return
	}

	query, execArgs, err := db.tpl.GetSQLContext(sql, dbParam)
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
