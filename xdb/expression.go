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
	GetSymbol() Symbol
	Build(state SqlState, input DBParam) (string, MissError)
}

// 表达式回调
type ExpressionBuildCallback func(item ExpressionValuer, state SqlState, param DBParam) (expression string, err MissError)

type ExpressionItem struct {
	Matcher                 ExpressionMatcher
	FullField               string
	PropName                string
	Oper                    string
	Symbol                  Symbol
	ExpressionBuildCallback ExpressionBuildCallback
}

func (m *ExpressionItem) GetSymbol() Symbol {
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

func (m *ExpressionItem) GetMatcher() ExpressionMatcher {
	return m.Matcher
}

func (m *ExpressionItem) Build(state SqlState, param DBParam) (expression string, err MissError) {
	if m.ExpressionBuildCallback == nil {
		return
	}
	state.SetDynamic(m.GetSymbol().DynamicType())
	return m.ExpressionBuildCallback(m, state, param)
}

func (m *ExpressionItem) GetOperatorCallback() (callback OperatorCallback, ok bool) {
	operator, ok := m.Matcher.GetOperatorMap().Load(m.Oper)
	if !ok {
		return nil, false
	}
	return operator.Callback, true
}
