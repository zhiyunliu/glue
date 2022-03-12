package dbs

import (
	"database/sql"
	"encoding/json"

	"github.com/zhiyunliu/velocity/extlib/xtypes"
)

type Result = sql.Result

//QueryRow 单行数据
type Row = xtypes.XMap

type Rows = xtypes.XMaps

//NewQueryRow 构建QueryRow对象
func NewRow() Row {
	return make(xtypes.XMap)
}

//NewQueryRowsByJSON 根据json创建QueryRows
func NewRowsByJSON(j string) (Rows, error) {
	rows := make([]xtypes.XMap, 0)
	err := json.Unmarshal([]byte(j), &rows)
	return rows, err
}

//NewQueryRows 构建QueryRows
func NewRows() Rows {
	return make([]xtypes.XMap, 0)
}
