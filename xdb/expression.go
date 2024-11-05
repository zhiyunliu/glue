package xdb

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/zhiyunliu/golibs/xtypes"
	//	"github.com/emirpasic/gods/v2/maps/treemap"
)

// 新建一个模板匹配器
var NewTemplateMatcher func(symbolMap SymbolMap, matchers ...ExpressionMatcher) TemplateMatcher

func init() {
	NewTemplateMatcher = NewDefaultTemplateMatcher
}

// 表达式解析选项
type ExpressionOptions struct {
	UseCache bool
}

type PropertyOption func(*ExpressionOptions)

// 使用解析缓存
func WithUseCache(use bool) PropertyOption {
	return func(o *ExpressionOptions) {
		o.UseCache = use
	}
}

// 属性表达式匹配器
type ExpressionMatcher interface {
	Name() string
	Pattern() string
	Symbol() SymbolMap
	MatchString(string) (ExpressionValuer, bool)
}

type ExpressionMatcherMap interface {
	Regist(...ExpressionMatcher)
	Load(name string) (ExpressionMatcher, bool)
	Each(call func(name string, matcher ExpressionMatcher) bool)
	Delete(name string)
	//	Clone() ExpressionMatcherMap
	BuildFullRegexp() *regexp.Regexp
}

// xdb表达式
type ExpressionValuer interface {
	GetPropName() string
	GetFullfield() string
	GetOper() string
	GetSymbol() string
	Build(input DBParam, argName string) (string, MissError)
}

// 表达式回调
type ExpressionBuildCallback func(item *ExpressionItem, param DBParam, argName string) (expression string, err MissError)

type ExpressionItem struct {
	FullField               string
	PropName                string
	Oper                    string
	Symbol                  string
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

func (m *ExpressionItem) Build(param DBParam, argName string) (expression string, err MissError) {
	if m.ExpressionBuildCallback == nil {
		return
	}
	return m.ExpressionBuildCallback(m, param, argName)
}

// PropertyPatternFunc
type PropertyPatternFunc func(matcherMap ExpressionMatcherMap) *regexp.Regexp

type TemplateMatcher interface {
	RegistMatcher(matcher ...ExpressionMatcher)
	GenerateSQL(item *SqlScene, sqlTpl string, input DBParam) (sql string, err error)
}

type DefaultTemplateMatcher struct {
	matcherMap  ExpressionMatcherMap
	exprCache   *sync.Map
	symbolMap   SymbolMap
	PatternFunc PropertyPatternFunc
}

// 新建一个默认模板匹配器
func NewDefaultTemplateMatcher(symbolMap SymbolMap, matchers ...ExpressionMatcher) TemplateMatcher {

	conn := &DefaultTemplateMatcher{
		matcherMap: NewExpressionMatcherMap(),
		exprCache:  &sync.Map{},
		symbolMap:  symbolMap,
	}
	conn.RegistMatcher(matchers...)
	return conn
}

func (conn *DefaultTemplateMatcher) RegistMatcher(matchers ...ExpressionMatcher) {
	conn.matcherMap.Regist(matchers...)
}

func (conn *DefaultTemplateMatcher) GenerateSQL(item *SqlScene, sqlTpl string, input DBParam) (sql string, err error) {

	matcherMap := conn.matcherMap
	word := matcherMap.BuildFullRegexp()

	var outerrs []MissError

	//@变量, 将数据放入params中
	sql = word.ReplaceAllStringFunc(sqlTpl, func(expr string) (repExpr string) {
		var (
			valuer ExpressionValuer
			ok     bool
		)
		matcherMap.Each(func(name string, matcher ExpressionMatcher) bool {
			valuer, ok = matcher.MatchString(expr)
			return !ok
		})

		symbol, ok := conn.symbolMap.Load(valuer.GetSymbol())
		if !ok {
			return expr
		}
		repExpr, err := symbol.Callback(item, valuer, input)
		if err != nil {
			outerrs = append(outerrs, err)
		}
		return

	})
	if len(outerrs) > 0 {
		return sql, NewMissListError(outerrs...)
	}
	return sql, nil
}

type DefaultExpressionMatcherMapImpl struct {
	mutex        *sync.Mutex
	matcherCache *sync.Map
	sortVal      map[string]int
	regexp       *atomic.Value
}

func NewExpressionMatcherMap() ExpressionMatcherMap {
	return &DefaultExpressionMatcherMapImpl{
		mutex:        &sync.Mutex{},
		matcherCache: &sync.Map{},
		sortVal:      map[string]int{},
		regexp:       &atomic.Value{},
	}
}

func (m *DefaultExpressionMatcherMapImpl) Regist(matchers ...ExpressionMatcher) {
	if len(matchers) <= 0 {
		return
	}
	for i := range matchers {
		matcher := matchers[i]
		if matcher == nil {
			continue
		}
		m.sortMatcher(matcher)
		m.matcherCache.Store(matcher.Name(), matcher)
		m.regexp.Store(nil)
	}
}
func (m *DefaultExpressionMatcherMapImpl) Load(name string) (ExpressionMatcher, bool) {
	tmp, ok := m.matcherCache.Load(name)
	if !ok {
		return nil, ok
	}
	return tmp.(ExpressionMatcher), ok
}
func (m *DefaultExpressionMatcherMapImpl) Each(call func(name string, matcher ExpressionMatcher) bool) {
	m.matcherCache.Range(func(key, value interface{}) bool {
		return call(key.(string), value.(ExpressionMatcher))
	})
}
func (m *DefaultExpressionMatcherMapImpl) Delete(name string) {
	m.matcherCache.Delete(name)
}

func (m *DefaultExpressionMatcherMapImpl) Clone() ExpressionMatcherMap {
	clone := NewExpressionMatcherMap()

	m.Each(func(name string, matcher ExpressionMatcher) bool {
		clone.Regist(matcher)
		return true
	})

	return clone
}

func (m *DefaultExpressionMatcherMapImpl) BuildFullRegexp() *regexp.Regexp {
	tmpVal := m.regexp.Load()
	if tmpVal != nil {
		return tmpVal.(*regexp.Regexp)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	tmpVal = m.regexp.Load()
	if tmpVal != nil {
		return tmpVal.(*regexp.Regexp)
	}

	sortMap := xtypes.NewSortedMap[int, string](func(a, b int) bool {
		return a < b
	})

	for k, v := range m.sortVal {
		sortMap.Put(v, k)
	}

	buffer := bytes.Buffer{}

	sortMap.Each(func(i int, matcherName string) {
		matcher, ok := m.Load(matcherName)
		if !ok {
			return
		}
		buffer.WriteString("(")
		buffer.WriteString(matcher.Pattern())
		buffer.WriteString(")|")
	})
	patternVal := buffer.String()
	patternVal = strings.TrimSuffix(patternVal, "|")

	pattern := regexp.MustCompile(patternVal)

	m.regexp.Store(pattern)
	return pattern
}

func (m *DefaultExpressionMatcherMapImpl) sortMatcher(matcher ExpressionMatcher) {
	_, ok := m.sortVal[matcher.Name()]
	if ok {
		return
	}
	m.sortVal[matcher.Name()] = len(m.sortVal)
}
