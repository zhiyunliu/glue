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

type emptyOperator struct {
}

func (o *emptyOperator) Name() string {
	return "empty"
}
func (o *emptyOperator) Callback(valuer xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
	return "empty"
}

func TestDefaultGetPropName(t *testing.T) {
	normalMatcher := NewNormalExpressionMatcher(DefaultSymbols, xdb.WithOperator(&emptyOperator{}))
	compareMatcher := NewCompareExpressionMatcher(DefaultSymbols, xdb.WithOperator(&emptyOperator{}))
	likeMatcher := NewLikeExpressionMatcher(DefaultSymbols, xdb.WithOperator(&emptyOperator{}))
	inMatcher := NewInExpressionMatcher(DefaultSymbols, xdb.WithOperator(&emptyOperator{}))

	tests := []struct {
		matcher       xdb.ExpressionMatcher
		name          string
		fullKey       string
		wantFullfield string
		wantPropName  string
		wantOper      string
		wantSymbol    string
		wantExpr      string
		wantErr       bool
		wantCanCache  bool
	}{

		{name: "1-1.", matcher: normalMatcher, fullKey: "@{field}", wantFullfield: "field", wantPropName: "field", wantOper: "@", wantSymbol: "@", wantExpr: "?", wantCanCache: true},
		{name: "1-2.", matcher: normalMatcher, fullKey: "${field}", wantFullfield: "field", wantPropName: "field", wantOper: "$", wantSymbol: "$", wantExpr: "f"},
		{name: "1-3.", matcher: normalMatcher, fullKey: "&{field}", wantFullfield: "field", wantPropName: "field", wantOper: "&", wantSymbol: "&", wantExpr: "and field=?"},
		{name: "1-4.", matcher: normalMatcher, fullKey: "|{field }", wantFullfield: "field", wantPropName: "field", wantOper: "|", wantSymbol: "|", wantExpr: "or field=?"},

		{name: "2-1.", matcher: normalMatcher, fullKey: "@{tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "@", wantSymbol: "@", wantExpr: "?", wantCanCache: true},
		{name: "2-2.", matcher: normalMatcher, fullKey: "${tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "$", wantSymbol: "$", wantExpr: "f"},
		{name: "2-3.", matcher: normalMatcher, fullKey: "&{tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "&", wantSymbol: "&", wantExpr: "and tbl.field=?"},
		{name: "2-4.", matcher: normalMatcher, fullKey: "|{tbl.field }", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "|", wantSymbol: "|", wantExpr: "or tbl.field=?"},

		{name: "4-1.", matcher: compareMatcher, fullKey: "&{> tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">", wantSymbol: "&", wantExpr: "and tbl.field>?"},
		{name: "5-1.", matcher: compareMatcher, fullKey: "&{>= tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">=", wantSymbol: "&", wantExpr: "and tbl.field>=?"},
		{name: "6-1.", matcher: compareMatcher, fullKey: "&{< tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<", wantSymbol: "&", wantExpr: "and tbl.field<?"},
		{name: "7-1.", matcher: compareMatcher, fullKey: "&{<= tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<=", wantSymbol: "&", wantExpr: "and tbl.field<=?"},

		{name: "4-2.", matcher: compareMatcher, fullKey: "|{> tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">", wantSymbol: "|", wantExpr: "or tbl.field>?"},
		{name: "5-3.", matcher: compareMatcher, fullKey: "|{>= tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">=", wantSymbol: "|", wantExpr: "or tbl.field>=?"},
		{name: "6-4.", matcher: compareMatcher, fullKey: "|{< tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<", wantSymbol: "|", wantExpr: "or tbl.field<?"},
		{name: "7-5.", matcher: compareMatcher, fullKey: "|{<= tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<=", wantSymbol: "|", wantExpr: "or tbl.field<=?"},

		{name: "b.", matcher: compareMatcher, fullKey: "&{>    tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">", wantSymbol: "&", wantExpr: "and tbl.field>?"},
		{name: "c.", matcher: compareMatcher, fullKey: "&{>=    tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">=", wantSymbol: "&", wantExpr: "and tbl.field>=?"},
		{name: "d.", matcher: compareMatcher, fullKey: "&{<    tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<", wantSymbol: "&", wantExpr: "and tbl.field<?"},
		{name: "e.", matcher: compareMatcher, fullKey: "&{<=    tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<=", wantSymbol: "&", wantExpr: "and tbl.field<=?"},

		{name: "0b.", matcher: compareMatcher, fullKey: "|{>tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">", wantSymbol: "|", wantExpr: "or tbl.field>?"},
		{name: "0c.", matcher: compareMatcher, fullKey: "|{>=tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: ">=", wantSymbol: "|", wantExpr: "or tbl.field>=?"},
		{name: "0d.", matcher: compareMatcher, fullKey: "|{<tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<", wantSymbol: "|", wantExpr: "or tbl.field<?"},
		{name: "0e.", matcher: compareMatcher, fullKey: "|{<=tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "<=", wantSymbol: "|", wantExpr: "or tbl.field<=?"},

		{name: "01b.", matcher: compareMatcher, fullKey: "|{>field}", wantFullfield: "field", wantPropName: "field", wantOper: ">", wantSymbol: "|", wantExpr: "or field>?"},
		{name: "01c.", matcher: compareMatcher, fullKey: "|{>=field}", wantFullfield: "field", wantPropName: "field", wantOper: ">=", wantSymbol: "|", wantExpr: "or field>=?"},
		{name: "01d.", matcher: compareMatcher, fullKey: "|{<field}", wantFullfield: "field", wantPropName: "field", wantOper: "<", wantSymbol: "|", wantExpr: "or field<?"},
		{name: "01e.", matcher: compareMatcher, fullKey: "|{<=field}", wantFullfield: "field", wantPropName: "field", wantOper: "<=", wantSymbol: "|", wantExpr: "or field<=?"},
		{name: "0b.", matcher: compareMatcher, fullKey: "&{<>   field}", wantFullfield: "field", wantPropName: "field", wantOper: "<>", wantSymbol: "&", wantExpr: "and field<>?"},
		{name: "1b.", matcher: compareMatcher, fullKey: "&{>   field}", wantFullfield: "field", wantPropName: "field", wantOper: ">", wantSymbol: "&", wantExpr: "and field>?"},
		{name: "1c.", matcher: compareMatcher, fullKey: "&{>=   field}", wantFullfield: "field", wantPropName: "field", wantOper: ">=", wantSymbol: "&", wantExpr: "and field>=?"},
		{name: "1d.", matcher: compareMatcher, fullKey: "&{<   field}", wantFullfield: "field", wantPropName: "field", wantOper: "<", wantSymbol: "&", wantExpr: "and field<?"},
		{name: "1e.", matcher: compareMatcher, fullKey: "&{<=   t.field}", wantFullfield: "t.field", wantPropName: "field", wantOper: "<=", wantSymbol: "&", wantExpr: "and t.field<=?"},
		{name: "0f.", matcher: compareMatcher, fullKey: "&{field <>   property}", wantFullfield: "field", wantPropName: "property", wantOper: "<>", wantSymbol: "&", wantExpr: "and field<>?"},
		{name: "1f.", matcher: compareMatcher, fullKey: "&{field >   property}", wantFullfield: "field", wantPropName: "property", wantOper: ">", wantSymbol: "&", wantExpr: "and field>?"},
		{name: "1g.", matcher: compareMatcher, fullKey: "&{t.field >=   property}", wantFullfield: "t.field", wantPropName: "property", wantOper: ">=", wantSymbol: "&", wantExpr: "and t.field>=?"},
		{name: "1h.", matcher: compareMatcher, fullKey: "&{t.field <   property}", wantFullfield: "t.field", wantPropName: "property", wantOper: "<", wantSymbol: "&", wantExpr: "and t.field<?"},
		{name: "1i.", matcher: compareMatcher, fullKey: "&{field <=   property}", wantFullfield: "field", wantPropName: "property", wantOper: "<=", wantSymbol: "&", wantExpr: "and field<=?"},
		{name: "1j.", matcher: compareMatcher, fullKey: "&{t.field<property}", wantFullfield: "t.field", wantPropName: "property", wantOper: "<", wantSymbol: "&", wantExpr: "and t.field<?"},
		{name: "1k.", matcher: compareMatcher, fullKey: "&{field<=property}", wantFullfield: "field", wantPropName: "property", wantOper: "<=", wantSymbol: "&", wantExpr: "and field<=?"},

		{name: "1a.", matcher: likeMatcher, fullKey: "&{like   field}", wantFullfield: "field", wantPropName: "field", wantOper: "like", wantSymbol: "&", wantExpr: "and field like ?"},
		{name: "3.", matcher: likeMatcher, fullKey: "&{like tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like", wantSymbol: "&", wantExpr: "and tbl.field like ?"},
		{name: "a.", matcher: likeMatcher, fullKey: "&{like    tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like", wantSymbol: "&", wantExpr: "and tbl.field like ?"},
		{name: "2a.", matcher: likeMatcher, fullKey: "&{like   %field}", wantFullfield: "field", wantPropName: "field", wantOper: "%like", wantSymbol: "&", wantExpr: "and field like '%'+?"},
		{name: "3a.", matcher: likeMatcher, fullKey: "&{like   field%}", wantFullfield: "field", wantPropName: "field", wantOper: "like%", wantSymbol: "&", wantExpr: "and field like ?+'%'"},
		{name: "4a.", matcher: likeMatcher, fullKey: "&{like   %field%}", wantFullfield: "field", wantPropName: "field", wantOper: "%like%", wantSymbol: "&", wantExpr: "and field like '%'+?+'%'"},
		{name: "5a.", matcher: likeMatcher, fullKey: "&{like    %tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like", wantSymbol: "&", wantExpr: "and tbl.field like '%'+?"},
		{name: "6a.", matcher: likeMatcher, fullKey: "&{like    tbl.field%}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like%", wantSymbol: "&", wantExpr: "and tbl.field like ?+'%'"},
		{name: "7a.", matcher: likeMatcher, fullKey: "&{like    %tbl.field%}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like%", wantSymbol: "&", wantExpr: "and tbl.field like '%'+?+'%'"},
		{name: "8a.", matcher: likeMatcher, fullKey: "&{like %tbl.field}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like", wantSymbol: "&", wantExpr: "and tbl.field like '%'+?"},
		{name: "9a.", matcher: likeMatcher, fullKey: "&{like tbl.field%}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "like%", wantSymbol: "&", wantExpr: "and tbl.field like ?+'%'"},
		{name: "10a.", matcher: likeMatcher, fullKey: "&{like %tbl.field%}", wantFullfield: "tbl.field", wantPropName: "field", wantOper: "%like%", wantSymbol: "&", wantExpr: "and tbl.field like '%'+?+'%'"},
		{name: "2b.", matcher: likeMatcher, fullKey: "&{tbl.field   like   %property}", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "%like", wantSymbol: "&", wantExpr: "and tbl.field like '%'+?"},
		{name: "3b.", matcher: likeMatcher, fullKey: "&{tbl.field   like   property%}", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "like%", wantSymbol: "&", wantExpr: "and tbl.field like ?+'%'"},
		{name: "4b.", matcher: likeMatcher, fullKey: "&{tbl.field   like   %property%}", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "%like%", wantSymbol: "&", wantExpr: "and tbl.field like '%'+?+'%'"},
		{name: "5b.", matcher: likeMatcher, fullKey: "&{field like    %property}", wantFullfield: "field", wantPropName: "property", wantOper: "%like", wantSymbol: "&", wantExpr: "and field like '%'+?"},
		{name: "6b.", matcher: likeMatcher, fullKey: "&{field like    property%}", wantFullfield: "field", wantPropName: "property", wantOper: "like%", wantSymbol: "&", wantExpr: "and field like ?+'%'"},
		{name: "7b.", matcher: likeMatcher, fullKey: "&{field like    %property%}", wantFullfield: "field", wantPropName: "property", wantOper: "%like%", wantSymbol: "&", wantExpr: "and field like '%'+?+'%'"},
		{name: "8b.", matcher: likeMatcher, fullKey: "&{tbl.field like %property}", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "%like", wantSymbol: "&", wantExpr: "and tbl.field like '%'+?"},
		{name: "9b.", matcher: likeMatcher, fullKey: "&{tbl.field like property%}", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "like%", wantSymbol: "&", wantExpr: "and tbl.field like ?+'%'"},
		{name: "10b.", matcher: likeMatcher, fullKey: "&{tbl.field like %property%}", wantFullfield: "tbl.field", wantPropName: "property", wantOper: "%like%", wantSymbol: "&", wantExpr: "and tbl.field like '%'+?+'%'"},

		{name: "8.", matcher: inMatcher, fullKey: "&{in tbl.infield}", wantFullfield: "tbl.infield", wantPropName: "infield", wantOper: "in", wantSymbol: "&", wantExpr: "and tbl.infield in (1,2)"},
		{name: "9.", matcher: inMatcher, fullKey: "&{in infield}", wantFullfield: "infield", wantPropName: "infield", wantOper: "in", wantSymbol: "&", wantExpr: "and infield in (1,2)"},
		{name: "f.", matcher: inMatcher, fullKey: "&{in    tbl.infield}", wantFullfield: "tbl.infield", wantPropName: "infield", wantOper: "in", wantSymbol: "&", wantExpr: "and tbl.infield in (1,2)"},
		{name: "g.", matcher: inMatcher, fullKey: "&{in    infield}", wantFullfield: "infield", wantPropName: "infield", wantOper: "in", wantSymbol: "&", wantExpr: "and infield in (1,2)"},
		{name: "h.", matcher: inMatcher, fullKey: "&{infield  in  inproperty}", wantFullfield: "infield", wantPropName: "inproperty", wantOper: "in", wantSymbol: "&", wantExpr: "and infield in ('p1','p2')"},
		{name: "i.", matcher: inMatcher, fullKey: "&{tt.infield  in    inproperty}", wantFullfield: "tt.infield", wantPropName: "inproperty", wantOper: "in", wantSymbol: "&", wantExpr: "and tt.infield in ('p1','p2')"},

		{name: "8-a.", matcher: inMatcher, fullKey: "&{in tbl.infield}", wantFullfield: "tbl.infield", wantPropName: "infield", wantOper: "in", wantSymbol: "&", wantExpr: "and tbl.infield in (1,2)"},
		{name: "9-a.", matcher: inMatcher, fullKey: "&{in infield}", wantFullfield: "infield", wantPropName: "infield", wantOper: "in", wantSymbol: "&", wantExpr: "and infield in (1,2)"},
		{name: "f-a.", matcher: inMatcher, fullKey: "&{in    tbl.infield}", wantFullfield: "tbl.infield", wantPropName: "infield", wantOper: "in", wantSymbol: "&", wantExpr: "and tbl.infield in (1,2)"},
		{name: "g-a.", matcher: inMatcher, fullKey: "&{in    infield}", wantFullfield: "infield", wantPropName: "infield", wantOper: "in", wantSymbol: "&", wantExpr: "and infield in (1,2)"},
		{name: "h-a.", matcher: inMatcher, fullKey: "&{infield  in   inproperty}", wantFullfield: "infield", wantPropName: "inproperty", wantOper: "in", wantSymbol: "&", wantExpr: "and infield in ('p1','p2')"},
		{name: "i-a.", matcher: inMatcher, fullKey: "&{tt.infield  in    inproperty}", wantFullfield: "tt.infield", wantPropName: "inproperty", wantOper: "in", wantSymbol: "&", wantExpr: "and tt.infield in ('p1','p2')"},
		{name: "j-a.", matcher: inMatcher, fullKey: "&{tt.infield  in    bytesfield}", wantFullfield: "tt.infield", wantPropName: "bytesfield", wantOper: "in", wantSymbol: "&", wantExpr: "", wantErr: true},
		{name: "j-a.", matcher: inMatcher, fullKey: "&{tt.infield  in    objfield}", wantFullfield: "tt.infield", wantPropName: "objfield", wantOper: "in", wantSymbol: "&", wantExpr: "", wantErr: true},

		{name: "$-array-1.", matcher: normalMatcher, fullKey: "${tbl.inproperty}", wantFullfield: "tbl.inproperty", wantPropName: "inproperty", wantOper: "$", wantSymbol: "$", wantExpr: "'p1','p2'"},
		{name: "$-array-2.", matcher: normalMatcher, fullKey: "${tbl.infield}", wantFullfield: "tbl.infield", wantPropName: "infield", wantOper: "$", wantSymbol: "$", wantExpr: "1,2"},

		{name: "@-empty-2.", matcher: normalMatcher, fullKey: "@{tbl.emptyfield}", wantFullfield: "tbl.emptyfield", wantPropName: "emptyfield", wantOper: "@", wantSymbol: "@", wantExpr: "?", wantCanCache: true},

		{name: "err-1.", matcher: normalMatcher, fullKey: "@{tbl.errfield}", wantFullfield: "tbl.errfield", wantPropName: "errfield", wantOper: "@", wantSymbol: "@", wantExpr: "", wantErr: true, wantCanCache: true},
	}

	dbParam := map[string]any{
		"property":   "p",
		"field":      "f",
		"inproperty": []string{"p1", "p2"},
		"infield":    []int{1, 2},
		"emptyfield": "",
		"bytesfield": []byte("bytes"),
		"objfield":   struct{}{},
	}

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
			if propValuer.GetSymbol().Name() != tt.wantSymbol {
				t.Errorf("GetSymbol() :%v, want %v", propValuer.GetSymbol(), tt.wantSymbol)
			}

			state := xdb.NewSqlState(&testPlaceHolder{prefix: "?"}, &xdb.TemplateOptions{UseExprCache: true})
			expr, err := propValuer.Build(state, dbParam)
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}

			if expr != tt.wantExpr {
				t.Errorf("Build() :%v, want %v", expr, tt.wantExpr)
			}

			if state.CanCache() != tt.wantCanCache {
				t.Errorf("Canche :%v, want:%v", state.CanCache(), tt.wantCanCache)
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
	symbolMap := DefaultSymbols
	symbolMap.Regist(&demoSymbols{})

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
			if got := getExpressionSymbol(symbolMap, tt.fullkey); got.Name() != tt.want {
				t.Errorf("getExpressionSymbol() = %v, want %v", got, tt.want)
			}
		})
	}
}

type demoSymbols struct{}

func (s *demoSymbols) Name() string {
	return "###"
}

func (s *demoSymbols) DynamicType() xdb.DynamicType {
	return xdb.DynamicNone
}

func (s *demoSymbols) Concat() string {
	return "demo"
}

func Benchmark_NormalMatcher(b *testing.B) {
	matcher := NewNormalExpressionMatcher(DefaultSymbols)
	tt := struct {
		name          string
		fullKey       string
		wantFullfield string
		wantPropName  string
		wantOper      string
		wantSymbol    string
		wantExpr      string
		wantErr       bool
		wantCanCache  bool
	}{
		name:          "1",
		fullKey:       `@{t.property}`,
		wantFullfield: `t.property`,
		wantPropName:  `property`,
		wantOper:      `@`,
		wantSymbol:    `@`,
		wantExpr:      `?`,
		wantErr:       false,
		wantCanCache:  true,
	}

	dbParam := map[string]interface{}{
		"property":   "p",
		"field":      "f",
		"inproperty": []string{"p1", "p2"},
		"infield":    []int{1, 2},
		"emptyfield": "",
		"bytesfield": []byte("bytes"),
		"objfield":   struct{}{},
	}

	for i := 0; i < b.N; i++ {

		propValuer, ok := matcher.MatchString(tt.fullKey)
		if !ok {
			b.Error("propValuer is null", tt.name)
			return
		}
		gotPropName := propValuer.GetPropName()

		if propValuer.GetFullfield() != tt.wantFullfield {
			b.Errorf("GetFullfield() :%v, want %v", propValuer.GetFullfield(), tt.wantFullfield)
		}
		if gotPropName != tt.wantPropName {
			b.Errorf("GetPropName() :%v, want %v", gotPropName, tt.wantPropName)
		}
		if propValuer.GetOper() != tt.wantOper {
			b.Errorf("GetOper() :%v, want %v", propValuer.GetOper(), tt.wantOper)
		}
		if propValuer.GetSymbol().Name() != tt.wantSymbol {
			b.Errorf("GetSymbol() :%v, want %v", propValuer.GetSymbol(), tt.wantSymbol)
		}

		state := xdb.NewSqlState(&testPlaceHolder{prefix: "?"}, &xdb.TemplateOptions{UseExprCache: true})
		expr, err := propValuer.Build(state, dbParam)
		if (err != nil) != tt.wantErr {
			b.Error(err)
		}

		if expr != tt.wantExpr {
			b.Errorf("Build() :%v, want %v", expr, tt.wantExpr)
		}

		if state.CanCache() != tt.wantCanCache {
			b.Errorf("Canche :%v, want:%v", state.CanCache(), tt.wantCanCache)
		}

	}
}
