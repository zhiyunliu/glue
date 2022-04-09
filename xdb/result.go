package xdb

import (
	"database/sql"

	"github.com/zhiyunliu/golibs/xtypes"
)

type Result = sql.Result

//Row 单行数据
type Row = xtypes.XMap

//Rows 多行数据
type Rows = xtypes.XMaps

//NewRow 构建Row对象
func NewRow() Row {
	return make(xtypes.XMap)
}

//NewRows 构建Rows
func NewRows() Rows {
	return make([]xtypes.XMap, 0)
}
