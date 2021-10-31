package dbs

import (
	"database/sql"
	"encoding/json"

	"github.com/zhiyunliu/velocity/libs/types"
)

type Result = sql.Result

//QueryRow 单行数据
type Row = types.XMap

type Rows = types.XMaps

//NewQueryRow 构建QueryRow对象
func NewRow() Row {
	return make(types.XMap)
}

//NewQueryRowsByJSON 根据json创建QueryRows
func NewRowsByJSON(j string) (Rows, error) {
	rows := make([]types.XMap, 0)
	err := json.Unmarshal([]byte(j), &rows)
	return rows, err
}

//NewQueryRows 构建QueryRows
func NewRows() Rows {
	return make([]types.XMap, 0)
}
