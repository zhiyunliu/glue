package tpl

import "testing"

func TestDefaultGetPropName(t *testing.T) {
	//field, tbl.field , tbl.field like , tbl.field >=
	tests := []struct {
		name          string
		fullKey       string
		wantFullfield string
		wantPropName  string
		wantOper      string
	}{
		{name: "1.", fullKey: "field", wantFullfield: "field", wantPropName: "field", wantOper: "="},
		{name: "2.", fullKey: "tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "="},
		{name: "3.", fullKey: "like tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like"},
		{name: "4.", fullKey: "> tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">"},
		{name: "5.", fullKey: ">= tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">="},
		{name: "6.", fullKey: "< tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<"},
		{name: "7.", fullKey: "<= tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<="},

		{name: "a.", fullKey: "like    tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like"},
		{name: "b.", fullKey: ">    tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">"},
		{name: "c.", fullKey: ">=    tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">="},
		{name: "d.", fullKey: "<    tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<"},
		{name: "e.", fullKey: "<=    tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<="},

		{name: "0b.", fullKey: ">tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">"},
		{name: "0c.", fullKey: ">=tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">="},
		{name: "0d.", fullKey: "<tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<"},
		{name: "0e.", fullKey: "<=tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<="},

		{name: "01b.", fullKey: ">field", wantFullfield: "field", wantPropName: "field", wantOper: ">"},
		{name: "01c.", fullKey: ">=field", wantFullfield: "field", wantPropName: "field", wantOper: ">="},
		{name: "01d.", fullKey: "<field", wantFullfield: "field", wantPropName: "field", wantOper: "<"},
		{name: "01e.", fullKey: "<=field", wantFullfield: "field", wantPropName: "field", wantOper: "<="},

		{name: "1a.", fullKey: "like   field", wantFullfield: "field", wantPropName: "field", wantOper: "like"},
		{name: "1b.", fullKey: ">   field", wantFullfield: "field", wantPropName: "field", wantOper: ">"},
		{name: "1c.", fullKey: ">=   field", wantFullfield: "field", wantPropName: "field", wantOper: ">="},
		{name: "1d.", fullKey: "<   field", wantFullfield: "field", wantPropName: "field", wantOper: "<"},
		{name: "1e.", fullKey: "<=   field", wantFullfield: "field", wantPropName: "field", wantOper: "<="},

		{name: "2a.", fullKey: "like   %field", wantFullfield: "field", wantPropName: "field", wantOper: "%like"},
		{name: "3a.", fullKey: "like   field%", wantFullfield: "field", wantPropName: "field", wantOper: "like%"},
		{name: "4a.", fullKey: "like   %field%", wantFullfield: "field", wantPropName: "field", wantOper: "%like%"},
		{name: "5a.", fullKey: "like    %tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like"},
		{name: "6a.", fullKey: "like    tbl.field%", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like%"},
		{name: "7a.", fullKey: "like    %tbl.field%", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like%"},
		{name: "8a.", fullKey: "like %tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like"},
		{name: "9a.", fullKey: "like tbl.field%", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like%"},
		{name: "10a.", fullKey: "like %tbl.field%", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFullField, gotPropName, gotOper := DefaultGetPropName(tt.fullKey)
			if gotFullField != tt.wantFullfield {
				t.Errorf("GetPropName() gotFullField:%v, want %v", gotFullField, tt.wantFullfield)
			}
			if gotPropName != tt.wantPropName {
				t.Errorf("GetPropName() gotPropName:%v, want %v", gotPropName, tt.wantPropName)
			}
			if gotOper != tt.wantOper {
				t.Errorf("GetPropName() gotOper:%v, want %v", gotOper, tt.wantOper)
			}
		})
	}
}
