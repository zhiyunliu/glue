package expression

import (
	"testing"

	"github.com/zhiyunliu/glue/xdb"
)

type testPlaceHolder struct {
	prefix string
}

func (ph *testPlaceHolder) Get(propName string) (argName, phName string) {
	phName = ph.prefix
	argName = propName
	return
}
func (ph *testPlaceHolder) BuildArgVal(argName string, val interface{}) interface{} {
	return val
}

func (ph *testPlaceHolder) NamedArg(propName string) (phName string) {
	phName = ph.prefix
	return
}

func (ph *testPlaceHolder) Clone() xdb.Placeholder {
	return &testPlaceHolder{
		prefix: ph.prefix,
	}
}

func TestDefaultGetPropName(t *testing.T) {
	normalMatcher := NewNormalExpressionMatcher(DefaultSymbols)
	compareMatcher := NewCompareExpressionMatcher(DefaultSymbols)
	likeMatcher := NewLikeExpressionMatcher(DefaultSymbols)
	inMatcher := NewInExpressionMatcher(DefaultSymbols)

	tests := []struct {
		matcher       xdb.ExpressionMatcher
		name          string
		fullKey       string
		wantFullfield string
		wantPropName  string
		wantOper      string
		wantSymbol    string
		wantExpr      string
	}{
		{name: "1-1.", matcher: normalMatcher, fullKey: "@{field}", wantFullfield: "field", wantPropName: "field", wantOper: "=", wantSymbol: "@"},
		{name: "1-2.", matcher: normalMatcher, fullKey: "${field}", wantFullfield: "field", wantPropName: "field", wantOper: "=", wantSymbol: "$"},
		{name: "1-3.", matcher: normalMatcher, fullKey: "&{field}", wantFullfield: "field", wantPropName: "field", wantOper: "=", wantSymbol: "&"},
		{name: "1-4.", matcher: normalMatcher, fullKey: "|{field }", wantFullfield: "field", wantPropName: "field", wantOper: "=", wantSymbol: "|"},

		{name: "2-1.", matcher: normalMatcher, fullKey: "@{tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "=", wantSymbol: "@"},
		{name: "2-2.", matcher: normalMatcher, fullKey: "${tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "=", wantSymbol: "$"},
		{name: "2-3.", matcher: normalMatcher, fullKey: "&{tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "=", wantSymbol: "&"},
		{name: "2-4.", matcher: normalMatcher, fullKey: "|{tbl.field }", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "=", wantSymbol: "|"},

		{name: "4-1.", matcher: compareMatcher, fullKey: "&{> tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">", wantSymbol: "&"},
		{name: "5-1.", matcher: compareMatcher, fullKey: "&{>= tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">=", wantSymbol: "&"},
		{name: "6-1.", matcher: compareMatcher, fullKey: "&{< tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<", wantSymbol: "&"},
		{name: "7-1.", matcher: compareMatcher, fullKey: "&{<= tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<=", wantSymbol: "&"},

		{name: "4-2.", matcher: compareMatcher, fullKey: "|{> tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">", wantSymbol: "|"},
		{name: "5-3.", matcher: compareMatcher, fullKey: "|{>= tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">=", wantSymbol: "|"},
		{name: "6-4.", matcher: compareMatcher, fullKey: "|{< tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<", wantSymbol: "|"},
		{name: "7-5.", matcher: compareMatcher, fullKey: "|{<= tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<=", wantSymbol: "|"},

		{name: "b.", matcher: compareMatcher, fullKey: "&{>    tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">", wantSymbol: "&"},
		{name: "c.", matcher: compareMatcher, fullKey: "&{>=    tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">=", wantSymbol: "&"},
		{name: "d.", matcher: compareMatcher, fullKey: "&{<    tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<", wantSymbol: "&"},
		{name: "e.", matcher: compareMatcher, fullKey: "&{<=    tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<=", wantSymbol: "&"},

		{name: "0b.", matcher: compareMatcher, fullKey: "|{>tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">", wantSymbol: "|"},
		{name: "0c.", matcher: compareMatcher, fullKey: "|{>=tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">=", wantSymbol: "|"},
		{name: "0d.", matcher: compareMatcher, fullKey: "|{<tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<", wantSymbol: "|"},
		{name: "0e.", matcher: compareMatcher, fullKey: "|{<=tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<=", wantSymbol: "|"},

		{name: "01b.", matcher: compareMatcher, fullKey: "|{>field}", wantFullfield: "field", wantPropName: "field", wantOper: ">", wantSymbol: "|"},
		{name: "01c.", matcher: compareMatcher, fullKey: "|{>=field}", wantFullfield: "field", wantPropName: "field", wantOper: ">=", wantSymbol: "|"},
		{name: "01d.", matcher: compareMatcher, fullKey: "|{<field}", wantFullfield: "field", wantPropName: "field", wantOper: "<", wantSymbol: "|"},
		{name: "01e.", matcher: compareMatcher, fullKey: "|{<=field}", wantFullfield: "field", wantPropName: "field", wantOper: "<=", wantSymbol: "|"},

		{name: "1b.", matcher: compareMatcher, fullKey: "&{>   field}", wantFullfield: "field", wantPropName: "field", wantOper: ">", wantSymbol: "&"},
		{name: "1c.", matcher: compareMatcher, fullKey: "&{>=   field}", wantFullfield: "field", wantPropName: "field", wantOper: ">=", wantSymbol: "&"},
		{name: "1d.", matcher: compareMatcher, fullKey: "&{<   field}", wantFullfield: "field", wantPropName: "field", wantOper: "<", wantSymbol: "&"},
		{name: "1e.", matcher: compareMatcher, fullKey: "&{<=   t.field}", wantFullfield: "t.field", wantPropName: "field", wantOper: "<=", wantSymbol: "&"},
		{name: "1f.", matcher: compareMatcher, fullKey: "&{field >   property}", wantFullfield: "field", wantPropName: "property", wantOper: ">", wantSymbol: "&"},
		{name: "1g.", matcher: compareMatcher, fullKey: "&{t.field >=   property}", wantFullfield: "t.field", wantPropName: "property", wantOper: ">=", wantSymbol: "&"},
		{name: "1h.", matcher: compareMatcher, fullKey: "&{t.field <   property}", wantFullfield: "t.field", wantPropName: "property", wantOper: "<", wantSymbol: "&"},
		{name: "1i.", matcher: compareMatcher, fullKey: "&{field <=   property}", wantFullfield: "field", wantPropName: "property", wantOper: "<=", wantSymbol: "&"},
		{name: "1j.", matcher: compareMatcher, fullKey: "&{t.field<property}", wantFullfield: "t.field", wantPropName: "property", wantOper: "<", wantSymbol: "&"},
		{name: "1k.", matcher: compareMatcher, fullKey: "&{field<=property}", wantFullfield: "field", wantPropName: "property", wantOper: "<=", wantSymbol: "&"},

		{name: "1a.", matcher: likeMatcher, fullKey: "&{like   field}", wantFullfield: "field", wantPropName: "field", wantOper: "like", wantSymbol: "&"},
		{name: "3.", matcher: likeMatcher, fullKey: "&{like tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like", wantSymbol: "&"},
		{name: "a.", matcher: likeMatcher, fullKey: "&{like    tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like", wantSymbol: "&"},
		{name: "2a.", matcher: likeMatcher, fullKey: "&{like   %field}", wantFullfield: "field", wantPropName: "field", wantOper: "%like", wantSymbol: "&"},
		{name: "3a.", matcher: likeMatcher, fullKey: "&{like   field%}", wantFullfield: "field", wantPropName: "field", wantOper: "like%", wantSymbol: "&"},
		{name: "4a.", matcher: likeMatcher, fullKey: "&{like   %field%}", wantFullfield: "field", wantPropName: "field", wantOper: "%like%", wantSymbol: "&"},
		{name: "5a.", matcher: likeMatcher, fullKey: "&{like    %tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like", wantSymbol: "&"},
		{name: "6a.", matcher: likeMatcher, fullKey: "&{like    tbl.field%}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like%", wantSymbol: "&"},
		{name: "7a.", matcher: likeMatcher, fullKey: "&{like    %tbl.field%}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like%", wantSymbol: "&"},
		{name: "8a.", matcher: likeMatcher, fullKey: "&{like %tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like", wantSymbol: "&"},
		{name: "9a.", matcher: likeMatcher, fullKey: "&{like tbl.field%}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like%", wantSymbol: "&"},
		{name: "10a.", matcher: likeMatcher, fullKey: "&{like %tbl.field%}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like%", wantSymbol: "&"},
		{name: "2b.", matcher: likeMatcher, fullKey: "&{tbl.field   like   %property}", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "%like", wantSymbol: "&"},
		{name: "3b.", matcher: likeMatcher, fullKey: "&{tbl.field   like   property%}", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "like%", wantSymbol: "&"},
		{name: "4b.", matcher: likeMatcher, fullKey: "&{tbl.field   like   %property%}", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "%like%", wantSymbol: "&"},
		{name: "5b.", matcher: likeMatcher, fullKey: "&{field like    %property}", wantFullfield: "field", wantPropName: "property", wantOper: "%like", wantSymbol: "&"},
		{name: "6b.", matcher: likeMatcher, fullKey: "&{field like    property%}", wantFullfield: "field", wantPropName: "property", wantOper: "like%", wantSymbol: "&"},
		{name: "7b.", matcher: likeMatcher, fullKey: "&{field like    %property%}", wantFullfield: "field", wantPropName: "property", wantOper: "%like%", wantSymbol: "&"},
		{name: "8b.", matcher: likeMatcher, fullKey: "&{tbl.field like %property}", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "%like", wantSymbol: "&"},
		{name: "9b.", matcher: likeMatcher, fullKey: "&{tbl.field like property%}", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "like%", wantSymbol: "&"},
		{name: "10b.", matcher: likeMatcher, fullKey: "&{tbl.field like %property%}", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "%like%", wantSymbol: "&"},

		{name: "8.", matcher: inMatcher, fullKey: "&{in tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "in", wantSymbol: "&"},
		{name: "9.", matcher: inMatcher, fullKey: "&{in field}", wantFullfield: "field", wantPropName: "field", wantOper: "in", wantSymbol: "&"},
		{name: "f.", matcher: inMatcher, fullKey: "&{in    tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "in", wantSymbol: "&"},
		{name: "g.", matcher: inMatcher, fullKey: "&{in    field}", wantFullfield: "field", wantPropName: "field", wantOper: "in", wantSymbol: "&"},
		{name: "h.", matcher: inMatcher, fullKey: "&{field  in   property}", wantFullfield: "field", wantPropName: "property", wantOper: "in", wantSymbol: "&"},
		{name: "i.", matcher: inMatcher, fullKey: "&{tt.field  in    property}", wantFullfield: "tt.field", wantPropName: "property", wantOper: "in", wantSymbol: "&"},

		{name: "8-a.", matcher: inMatcher, fullKey: "&{in tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "in", wantSymbol: "&"},
		{name: "9-a.", matcher: inMatcher, fullKey: "&{in field}", wantFullfield: "field", wantPropName: "field", wantOper: "in", wantSymbol: "&"},
		{name: "f-a.", matcher: inMatcher, fullKey: "&{in    tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "in", wantSymbol: "&"},
		{name: "g-a.", matcher: inMatcher, fullKey: "&{in    field}", wantFullfield: "field", wantPropName: "field", wantOper: "in", wantSymbol: "&"},
		{name: "h-a.", matcher: inMatcher, fullKey: "&{field  in   property}", wantFullfield: "field", wantPropName: "property", wantOper: "in", wantSymbol: "&"},
		{name: "i-a.", matcher: inMatcher, fullKey: "&{tt.field  in    property}", wantFullfield: "tt.field", wantPropName: "property", wantOper: "in", wantSymbol: "&"},
	}

	dbParam := map[string]any{"property": "p", "field": "field"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := tt.matcher
			propValuer, ok := matcher.MatchString(tt.fullKey)
			if !ok {
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
			if propValuer.GetSymbol() != tt.wantSymbol {
				t.Errorf("GetSymbol() :%v, want %v", propValuer.GetSymbol(), tt.wantSymbol)
			}
			state := xdb.NewDefaultSqlState(&testPlaceHolder{prefix: "test"}, &xdb.TemplateOptions{UseExprCache: true})
			expr, err := propValuer.Build(state, dbParam)
			if err != nil {
				t.Error(err)
			}

			if expr != tt.wantExpr {
				t.Errorf("Build() :%v, want %v", expr, tt.wantExpr)

			}
		})
	}
}

func Test_getExpressionPropertyName(t *testing.T) {

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
				t.Errorf("getExpressionPropertyName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getExpressionSymbol(t *testing.T) {

	tests := []struct {
		name    string
		fullkey string
		want    string
	}{
		{name: "1", fullkey: "@{aaa.bbb}", want: "@"},
		{name: "2", fullkey: "${bbb}", want: "$"},
		{name: "3", fullkey: "&{bbb}", want: "&"},
		{name: "4", fullkey: "|{bbb}", want: "|"},
		{name: "5", fullkey: "###{bbb}", want: "###"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getExpressionSymbol(tt.fullkey); got != tt.want {
				t.Errorf("getExpressionSymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}
