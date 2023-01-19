package tpl

import (
	"reflect"
	"testing"
	"time"

	"github.com/zhiyunliu/golibs/xtypes"
)

func TestDBParam_Get(t *testing.T) {

	var datetime interface{} = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	decimal := xtypes.NewDecimalFromInt(100)

	val := DBParam{
		"int":      1,
		"string":   "string",
		"datetime": datetime,
		"decimal":  decimal,
		"int32":    int32(32),
		"int64":    int64(64),
		"float32":  float32(32.1),
		"float64":  float64(64.2),
		"bool":     true,
		"[]byte":   []byte("[]byte"),
		"array":    []int{1, 2, 3},
		"array2":   []string{"1", "2", "3"},
	}

	tests := []struct {
		name    string
		key     string
		wantVal interface{}
	}{
		{name: "1.", key: "int", wantVal: 1},
		{name: "2.", key: "string", wantVal: "string"},
		{name: "3.", key: "datetime", wantVal: "2023-01-01 00:00:00"},
		{name: "4.", key: "decimal", wantVal: "100"},
		{name: "5.", key: "int32", wantVal: int32(32)},
		{name: "6.", key: "int64", wantVal: int64(64)},
		{name: "7.", key: "float32", wantVal: float32(32.1)},
		{name: "8.", key: "float64", wantVal: float64(64.2)},
		{name: "9.", key: "bool", wantVal: true},
		{name: "10.", key: "[]byte", wantVal: []byte("[]byte")},

		// {name: "1.", key: "array", wantVal: "1,2,3"},
		// {name: "1.", key: "array2", wantVal: "'1','2','3'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotVal := val.Get(tt.key); !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("DBParam.Get() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}
