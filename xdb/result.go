package xdb

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/zhiyunliu/golibs/bytesconv"
	"github.com/zhiyunliu/golibs/xtypes"
)

type Result = sql.Result

// Row 单行数据
type Row = xtypes.XMap

// Rows 多行数据
type Rows = xtypes.XMaps

// NewRow 构建Row对象
func NewRow() Row {
	return make(xtypes.XMap)
}

// NewRows 构建Rows
func NewRows() Rows {
	return make([]xtypes.XMap, 0)
}

type RawMessage []byte

// MarshalJSON returns m as the JSON encoding of m.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("xdb.RawMessage: UnmarshalJSON on nil pointer")
	}
	var orgdata string
	if err := json.Unmarshal(data, &orgdata); err != nil {
		//todo:这里可能会解析出本不想解析的内容
		orgdata = bytesconv.BytesToString(data)
	}

	*m = append((*m)[0:0], bytesconv.StringToBytes(orgdata)...)
	return nil
}

var _ json.Marshaler = (*RawMessage)(nil)
var _ json.Unmarshaler = (*RawMessage)(nil)

// DataRowReader 数据行读取接口
type DataRowReader interface {
	// GetNewRow 获取一个数据行对象
	GetNewRow() any
	// PresentRow 填充数据行对象
	PresentRow(data any) error
	// GetNext 获取下一个数据行对象
	GetNext() (any, bool)
	// Close 关闭数据行读取器
	Close() error
}
