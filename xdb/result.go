package xdb

import (
	"database/sql"
	"encoding/json"

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

//NewRowsByJSON 根据json创建Rows
func NewRowsByJSON(j string) (Rows, error) {
	rows := make([]xtypes.XMap, 0)
	err := json.Unmarshal([]byte(j), &rows)
	return rows, err
}
