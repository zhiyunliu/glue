package xdb

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/zhiyunliu/glue/contrib/xdb/internal"
	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xtypes"
)

// DB 数据库操作类
type xDB struct {
	cfg *Setting
	db  internal.ISysDB
	tpl tpl.SQLTemplate
}

// NewDB 创建DB实例
func NewDB(proto string, setting *Setting) (obj xdb.IDB, err error) {
	newCfg, err := xdb.DefaultRefactor(setting.ConnName, setting.Cfg)
	if err != nil {
		return
	}
	if newCfg != nil {
		setting.Cfg = newCfg
	}
	conn := setting.Cfg.Conn
	maxOpen := setting.Cfg.MaxOpen
	maxIdle := setting.Cfg.MaxIdle
	maxLifeTime := setting.Cfg.LifeTime

	if maxOpen <= 0 {
		maxOpen = runtime.NumCPU() * 10
	}
	if maxIdle <= 0 {
		maxIdle = maxOpen
	}

	dbobj := &xDB{
		cfg: setting,
	}

	setting.slowThreshold = time.Duration(setting.Cfg.LongQueryTime) * time.Millisecond
	if setting.Cfg.LoggerName != "" {
		setting.logger, _ = xdb.GetLogger(setting.Cfg.LoggerName)
	}

	dbobj.tpl, err = tpl.GetDBTemplate(proto)
	if err != nil {
		return
	}
	dbobj.db, err = internal.NewSysDB(proto, conn, maxOpen, maxIdle, time.Duration(maxLifeTime)*time.Second)
	return dbobj, err
}
func (db *xDB) GetImpl() interface{} {
	return db.db
}

// Query 查询数据
func (db *xDB) Query(ctx context.Context, sql string, input map[string]interface{}) (rows xdb.Rows, err error) {
	start := time.Now()

	query, execArgs := db.tpl.GetSQLContext(sql, input)

	debugPrint(ctx, db.cfg, query, execArgs...)
	data, err := db.db.Query(query, execArgs...)
	if err != nil {
		return nil, internal.GetError(err, query, execArgs...)
	}
	defer func() {
		if data != nil {
			data.Close()
		}
	}()
	rows, err = internal.ResolveRows(data)
	if err != nil {
		return nil, internal.GetError(err, query, execArgs...)
	}
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)

	return
}

// Multi 查询数据(多个数据集)
func (db *xDB) Multi(ctx context.Context, sql string, input map[string]interface{}) (datasetRows []xdb.Rows, err error) {
	start := time.Now()

	query, execArgs := db.tpl.GetSQLContext(sql, input)

	debugPrint(ctx, db.cfg, query, execArgs...)
	sqlRows, err := db.db.Query(query, execArgs...)
	if err != nil {
		return nil, internal.GetError(err, query, execArgs...)
	}
	defer func() {
		if sqlRows != nil {
			sqlRows.Close()
		}
	}()
	datasetRows, err = internal.ResolveMultiRows(sqlRows)
	if err != nil {
		return nil, internal.GetError(err, query, execArgs...)
	}
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)

	return
}

func (db *xDB) First(ctx context.Context, sql string, input map[string]interface{}) (data xdb.Row, err error) {
	rows, err := db.Query(ctx, sql, input)
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

func (db *xDB) Scalar(ctx context.Context, sql string, input map[string]interface{}) (data interface{}, err error) {
	rows, err := db.Query(ctx, sql, input)
	if err != nil {
		return
	}
	if rows.Len() == 0 || len(rows[0]) == 0 {
		return nil, nil
	}
	data, _ = rows[0].Get(rows[0].Keys()[0])
	return
}

// Execute 根据包含@名称占位符的语句执行查询语句
func (db *xDB) Exec(ctx context.Context, sql string, input map[string]interface{}) (r xdb.Result, err error) {
	start := time.Now()

	query, execArgs := db.tpl.GetSQLContext(sql, input)

	debugPrint(ctx, db.cfg, query, execArgs...)
	r, err = db.db.Exec(query, execArgs...)
	if err != nil {
		return nil, internal.GetError(err, query, execArgs...)
	}
	printSlowQuery(ctx, db.cfg, time.Since(start), query, execArgs...)

	return
}

// Begin 创建事务
func (db *xDB) Begin() (t xdb.ITrans, err error) {
	tt := &xTrans{
		cfg: db.cfg,
	}
	tt.tx, err = db.db.Begin()
	if err != nil {
		return
	}
	tt.tpl = db.tpl
	return tt, nil
}

// Transaction 执行事务
func (db *xDB) Transaction(callback xdb.TransactionCallback) (err error) {
	tt := &xTrans{
		cfg: db.cfg,
	}
	tt.tx, err = db.db.Begin()
	if err != nil {
		return
	}
	tt.tpl = db.tpl
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

// Close  关闭当前数据库连接
func (db *xDB) Close() error {
	return db.db.Close()
}
