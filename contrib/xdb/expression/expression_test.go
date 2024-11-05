package expression

import (
	"testing"

	"github.com/zhiyunliu/glue/xdb"
)

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
		{name: "1f.", fullKey: "field >   property", wantFullfield: "field", wantPropName: "property", wantOper: ">"},
		{name: "1g.", fullKey: "t.field >=   property", wantFullfield: "t.field", wantPropName: "property", wantOper: ">="},
		{name: "1h.", fullKey: "t.field <   property", wantFullfield: "t.field", wantPropName: "property", wantOper: "<"},
		{name: "1i.", fullKey: "field <=   property", wantFullfield: "field", wantPropName: "property", wantOper: "<="},
		{name: "1j.", fullKey: "t.field<property", wantFullfield: "t.field", wantPropName: "property", wantOper: "<"},
		{name: "1k.", fullKey: "field<=property", wantFullfield: "field", wantPropName: "property", wantOper: "<="},

		{name: "2a.", fullKey: "like   %field", wantFullfield: "field", wantPropName: "field", wantOper: "%like"},
		{name: "3a.", fullKey: "like   field%", wantFullfield: "field", wantPropName: "field", wantOper: "like%"},
		{name: "4a.", fullKey: "like   %field%", wantFullfield: "field", wantPropName: "field", wantOper: "%like%"},
		{name: "5a.", fullKey: "like    %tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like"},
		{name: "6a.", fullKey: "like    tbl.field%", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like%"},
		{name: "7a.", fullKey: "like    %tbl.field%", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like%"},
		{name: "8a.", fullKey: "like %tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like"},
		{name: "9a.", fullKey: "like tbl.field%", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like%"},
		{name: "10a.", fullKey: "like %tbl.field%", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like%"},

		{name: "2b.", fullKey: "tbl.field   like   %property", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "%like"},
		{name: "3b.", fullKey: "tbl.field   like   property%", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "like%"},
		{name: "4b.", fullKey: "tbl.field   like   %property%", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "%like%"},
		{name: "5b.", fullKey: "field like    %property", wantFullfield: "field", wantPropName: "property", wantOper: "%like"},
		{name: "6b.", fullKey: "field like    property%", wantFullfield: "field", wantPropName: "property", wantOper: "like%"},
		{name: "7b.", fullKey: "field like    %property%", wantFullfield: "field", wantPropName: "property", wantOper: "%like%"},
		{name: "8b.", fullKey: "tbl.field like %property", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "%like"},
		{name: "9b.", fullKey: "tbl.field like property%", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "like%"},
		{name: "10b.", fullKey: "tbl.field like %property%", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "%like%"},

		{name: "8.", fullKey: "in tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "in"},
		{name: "9.", fullKey: "in field", wantFullfield: "field", wantPropName: "field", wantOper: "in"},
		{name: "f.", fullKey: "in    tbl.field", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "in"},
		{name: "g.", fullKey: "in    field", wantFullfield: "field", wantPropName: "field", wantOper: "in"},
		{name: "h.", fullKey: "field  in   property", wantFullfield: "field", wantPropName: "property", wantOper: "in"},
		{name: "i.", fullKey: "tt.field  in    property", wantFullfield: "tt.field", wantPropName: "property", wantOper: "in"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			propValuer := GetExpressionValuer(tt.fullKey, &xdb.ExpressionOptions{UseCache: true})
			if propValuer == nil {
				t.Error("propValuer is null", tt.name)
				return
			}
			gotPropName := propValuer.GetPropName()

			if propValuer.GetFullfield() != tt.wantFullfield {
				t.Errorf("GetFullfield() :%v, want %v", propValuer.GetFullfield(), tt.wantFullfield)
			}
			if gotPropName != tt.wantPropName {
				t.Errorf("GetPropName() :%v, want %v", gotPropName, tt.wantPropName)
			}
			if propValuer.GetOper() != tt.wantOper {
				t.Errorf("GetOper() :%v, want %v", propValuer.GetOper(), tt.wantOper)
			}
		})
	}
}

func Test_getPropertyName(t *testing.T) {

	tests := []struct {
		name    string
		fullkey string
		want    string
	}{
		{name: "1", fullkey: "aaa.bbb", want: "bbb"},
		{name: "2", fullkey: "bbb", want: "bbb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getExpressionPropertyName(tt.fullkey); got != tt.want {
				t.Errorf("getPropertyName() = %v, want %v", got, tt.want)
			}
		})
	}
}
