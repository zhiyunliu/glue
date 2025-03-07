package tpl

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/zhiyunliu/glue/xdb"
)

func TestDefaultTemplateMatcher_GenerateSQL(t *testing.T) {

	var conn xdb.TemplateMatcher = NewDefaultTemplateMatcher(&testExpressionMatcher{
		symbolMap: xdb.NewSymbolMap(&testSymbol{}, &test2Symbol{}),
	}, &test2ExpressionMatcher{
		symbolMap: xdb.NewSymbolMap(&testSymbol{}, &test2Symbol{}),
	})

	tests := []struct {
		name     string
		sqlTpl   string
		input    xdb.DBParam
		wantVals []any
		wantSql  string
		wantErr  bool
	}{
		{name: "1.", sqlTpl: "select 1 from t where t.aa=@{aa}", input: map[string]any{"aa": 1}, wantSql: "select 1 from t where t.aa=@p_aa", wantVals: []any{1}, wantErr: false},
		{name: "1-cache.", sqlTpl: "select 1 from t where t.aa=@{aa}", input: map[string]any{"aa": 2}, wantSql: "select 1 from t where t.aa=@p_aa", wantVals: []any{2}, wantErr: false},
		{name: "2.", sqlTpl: "select 1 from t where t.aa=@{aa} and t.bb=@{t.bb}", input: map[string]any{"aa": 1, "bb": "b"}, wantSql: "select 1 from t where t.aa=@p_aa and t.bb=@p_bb", wantVals: []any{1, "b"}, wantErr: false},
		{name: "3.error", sqlTpl: "select 1 from t where t.aa=@{aa} and t.bb=@{t.abb}", input: map[string]any{"aa": 1, "bb": "b"}, wantSql: "select 1 from t where t.aa=@p_aa and t.bb=@{t.abb}", wantVals: []any{1}, wantErr: true},
		{name: "3-nocache", sqlTpl: "select 1 from t where t.aa=@{aa} &{t.aabb=bb}", input: map[string]any{"aa": 1, "bb": "b"}, wantSql: "select 1 from t where t.aa=@p_aa and t.aabb=@p_bb", wantVals: []any{1, "b"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			state := xdb.NewSqlState(&testPlaceHolder{prefix: "p_"})
			state.WithTemplateOptions(&xdb.TemplateOptions{UseExprCache: true})

			gotSql, err := conn.GenerateSQL(state, tt.sqlTpl, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultTemplateMatcher.GenerateSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSql != tt.wantSql {
				t.Errorf("DefaultTemplateMatcher.GenerateSQL() = %v, want %v", gotSql, tt.wantSql)
			}

			if !reflect.DeepEqual(tt.wantVals, state.GetValues()) {
				t.Errorf("DefaultTemplateMatcher GetValues() got=%v wantVals=%v", state.GetValues(), tt.wantVals)
			}

		})
	}
}

type testExpressionMatcher struct {
	symbolMap xdb.SymbolMap
}

func (m *testExpressionMatcher) Name() string {
	return "test"
}

func (m *testExpressionMatcher) Pattern() string {
	const pattern = `@({(\w+(\.\w+)?\s*)})`
	return pattern
}

func (m *testExpressionMatcher) GetOperatorMap() xdb.OperatorMap {
	operList := []xdb.Operator{

		xdb.NewOperator("@", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return phName
		}),

		xdb.NewOperator("&", func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
			return fmt.Sprintf("%s %s=%s", item.GetSymbol().Concat(), item.GetFullfield(), phName)
		}),
	}

	return xdb.NewOperatorMap(operList...)

}
func (m *testExpressionMatcher) MatchString(expression string) (valuer xdb.ExpressionValuer, ok bool) {

	ok = strings.HasPrefix(expression, "@")
	if !ok {
		return
	}
	expression = strings.Trim(expression, "@{}")
	expression = strings.TrimSpace(expression)

	fullkey := expression

	symbol, _ := m.symbolMap.Load("@")

	item := &xdb.ExpressionItem{
		Symbol:    symbol,
		Matcher:   m,
		FullField: fullkey,
		PropName:  fullkey,
	}
	item.Oper = item.Symbol.Name()
	pIdx := strings.Index(fullkey, ".")

	if pIdx > 0 {
		item.PropName = fullkey[pIdx+1:]
	}
	item.ExpressionBuildCallback = m.defaultBuildCallback()

	return item, ok
}
func (m *testExpressionMatcher) defaultBuildCallback() xdb.ExpressionBuildCallback {
	return func(item xdb.ExpressionValuer, state xdb.SqlState, param xdb.DBParam) (expression string, err xdb.MissError) {
		val, err := param.GetVal(item.GetPropName())
		if err != nil {
			return
		}
		phName := state.AppendExpr(item.GetPropName(), val)
		return phName, nil
	}
}

type test2ExpressionMatcher struct {
	symbolMap xdb.SymbolMap
}

func (m *test2ExpressionMatcher) Name() string {
	return "test2"
}

func (m *test2ExpressionMatcher) Pattern() string {
	const pattern = `[&|\|](({((\w+\.)?\w+)\s*(>|>=|<>|=|<|<=)\s*(\w+)})|({(>|>=|<>|=|<|<=)\s*(\w+(\.\w+)?)}))`
	return pattern
}

func (m *test2ExpressionMatcher) GetOperatorMap() xdb.OperatorMap {

	operCallback := func(item xdb.ExpressionValuer, param xdb.DBParam, phName string, value any) string {
		return fmt.Sprintf("%s %s%s%s", item.GetSymbol().Concat(), item.GetFullfield(), item.GetOper(), phName)
	}
	operList := []xdb.Operator{
		xdb.NewOperator(">", operCallback),
		xdb.NewOperator(">=", operCallback),
		xdb.NewOperator("<>", operCallback),
		xdb.NewOperator("=", operCallback),
		xdb.NewOperator("<", operCallback),
		xdb.NewOperator("<=", operCallback),
	}

	return xdb.NewOperatorMap(operList...)

}
func (m *test2ExpressionMatcher) MatchString(expression string) (valuer xdb.ExpressionValuer, ok bool) {
	ok = true

	expression = strings.Trim(expression, "&{}")
	expression = strings.TrimSpace(expression)

	parties := strings.Split(expression, "=")

	fullkey := parties[0]
	propName := parties[1]
	symbol, _ := m.symbolMap.Load("&")

	item := &xdb.ExpressionItem{
		Symbol:    symbol,
		Matcher:   m,
		FullField: fullkey,
		PropName:  propName,
	}
	item.Oper = "="

	item.ExpressionBuildCallback = m.defaultBuildCallback()

	return item, ok
}
func (m *test2ExpressionMatcher) defaultBuildCallback() xdb.ExpressionBuildCallback {
	return func(item xdb.ExpressionValuer, state xdb.SqlState, param xdb.DBParam) (expression string, err xdb.MissError) {
		val, err := param.GetVal(item.GetPropName())
		if err != nil {
			return
		}
		phName := state.AppendExpr(item.GetPropName(), val)

		return fmt.Sprintf("and %s%s%s", item.GetFullfield(), item.GetOper(), phName), nil
	}
}

type testPlaceHolder struct {
	prefix string
}

func (ph *testPlaceHolder) Get(propName string) (argName, phName string) {
	argName = fmt.Sprint(ph.prefix, propName)
	phName = ph.NamedArg(argName)
	return
}

func (ph *testPlaceHolder) NamedArg(argName string) (phName string) {
	phName = "@" + argName
	return
}

func (ph *testPlaceHolder) BuildArgVal(argName string, val any) any {
	if arg, ok := val.(sql.NamedArg); ok {
		return arg.Value
	}
	return val

}

type testSymbol struct{}

func (s *testSymbol) Name() string {
	return xdb.SymbolAt
}

func (s *testSymbol) DynamicType() xdb.DynamicType {
	return xdb.DynamicNone
}

func (s *testSymbol) Concat() string {
	return ""
}
func (s *testSymbol) IsDynamic() bool {
	return false
}

type test2Symbol struct{}

func (s *test2Symbol) Name() string {
	return xdb.SymbolAnd
}

func (s *test2Symbol) DynamicType() xdb.DynamicType {
	return xdb.DynamicAnd
}

func (s *test2Symbol) Concat() string {
	return ""
}
func (s *test2Symbol) IsDynamic() bool {
	return true
}
