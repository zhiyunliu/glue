package xdb

import (
	"regexp"
	//	"github.com/emirpasic/gods/v2/maps/treemap"
)

// 新建一个模板匹配器
var NewTemplateMatcher func(matchers ...ExpressionMatcher) TemplateMatcher

func init() {
	NewTemplateMatcher = NewDefaultTemplateMatcher
}

// 属性表达式匹配器
type ExpressionMatcher interface {
	Name() string
	Pattern() string
	GetOperatorMap() OperatorMap
	MatchString(string) (ExpressionValuer, bool)
}

type ExpressionMatcherMap interface {
	Regist(...ExpressionMatcher)
	Load(name string) (ExpressionMatcher, bool)
	Find(call func(matcher ExpressionMatcher) bool) ExpressionMatcher
	GetMatcherRegexp() *regexp.Regexp
}

// xdb表达式
type ExpressionValuer interface {
	GetMatcher() ExpressionMatcher
	GetOperatorCallback() (callback OperatorCallback, ok bool)
	GetPropName() string
	GetFullfield() string
	GetOper() string
	GetSymbol() string
	GetConcat() string
	Build(state SqlState, input DBParam) (string, MissError)
}

// 表达式回调
type ExpressionBuildCallback func(item ExpressionValuer, state SqlState, param DBParam) (expression string, err MissError)

type ExpressionItem struct {
	Matcher                 ExpressionMatcher
	FullField               string
	PropName                string
	Oper                    string
	Symbol                  string
	Concat                  string
	ExpressionBuildCallback ExpressionBuildCallback
}

func (m *ExpressionItem) GetSymbol() string {
	return m.Symbol
}

func (m *ExpressionItem) GetPropName() string {
	return m.PropName
}

func (m *ExpressionItem) GetFullfield() string {
	return m.FullField
}

func (m *ExpressionItem) GetOper() string {
	return m.Oper
}
func (m *ExpressionItem) GetConcat() string {
	return m.Concat
}
func (m *ExpressionItem) GetMatcher() ExpressionMatcher {
	return m.Matcher
}

func (m *ExpressionItem) Build(state SqlState, param DBParam) (expression string, err MissError) {
	if m.ExpressionBuildCallback == nil {
		return
	}
	return m.ExpressionBuildCallback(m, state, param)
}

func (m *ExpressionItem) GetOperatorCallback() (callback OperatorCallback, ok bool) {
	operator, ok := m.Matcher.GetOperatorMap().Load(m.Oper)
	if !ok {
		return nil, false
	}
	return operator.Callback, true
}

func (m *ExpressionItem) SpecConcat(symbolMap SymbolMap) {
	tmp, _ := symbolMap.Load(m.Symbol)
	m.Concat = tmp.Concat()
}
