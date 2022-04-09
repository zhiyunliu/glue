package main

import (
	"fmt"
	"time"

	"github.com/zhiyunliu/gel/context"
	"github.com/zhiyunliu/gel/xdb"
)

type DBdemo struct{}

func (d *DBdemo) QueryHandle(ctx context.Context) interface{} {
	dbobj := xdb.GetDB("localhost")
	idval := ctx.Request().Query().Get("id")
	sql := `select id,name from new_table t`
	if len(idval) > 0 {
		sql = `select id,name from new_table t where t.id=@id`
	}
	rows, err := dbobj.Query(sql, map[string]interface{}{
		"id": idval,
	})
	if err != nil {
		ctx.Log().Error(err)
	}

	return rows

}

func (d *DBdemo) FirstHandle(ctx context.Context) interface{} {
	dbobj := xdb.GetDB("localhost")
	row, err := dbobj.First("select id,name from new_table t where t.id=@id", map[string]interface{}{
		"id": ctx.Request().Query().Get("id"),
	})
	if err != nil {
		ctx.Log().Error(err)
	}

	return row

}

func (d *DBdemo) InsertHandle(ctx context.Context) interface{} {
	dbobj := xdb.GetDB("localhost")
	result, err := dbobj.Exec("insert into new_table(name) values(@name) ", map[string]interface{}{
		"name": fmt.Sprintf("insert:%s:%s", ctx.Request().Query().Get("name"), time.Now().Format("2006-01-02 15:04:05")),
	})
	if err != nil {
		ctx.Log().Error(err)
	}

	lastId, err1 := result.LastInsertId()
	effCnt, err2 := result.RowsAffected()
	return map[string]interface{}{
		"LastInsertId": lastId,
		"RowsAffected": effCnt,
		"Error1":       err1,
		"Error2":       err2,
	}
}

func (d *DBdemo) UpdateHandle(ctx context.Context) interface{} {
	dbobj := xdb.GetDB("localhost")
	result, err := dbobj.Exec("update new_table set name=@name where id=@id ", map[string]interface{}{
		"id":   ctx.Request().Query().Get("id"),
		"name": fmt.Sprintf("update:%s", time.Now().Format("2006-01-02 15:04:05")),
	})
	if err != nil {
		ctx.Log().Error(err)
	}
	lastId, err1 := result.LastInsertId()
	effCnt, err2 := result.RowsAffected()
	return map[string]interface{}{
		"LastInsertId": lastId,
		"RowsAffected": effCnt,
		"Error1":       err1,
		"Error2":       err2,
	}
}
