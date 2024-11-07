package xdb

import (
	"bytes"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/zhiyunliu/golibs/xtypes"
)

// 表达式解析选项
type TemplateOptions struct {
	UseCache bool
}

type TemplateOption func(*TemplateOptions)

// 使用解析缓存
func WithUseCache(use bool) TemplateOption {
	return func(o *TemplateOptions) {
		o.UseCache = use
	}
}

type TemplateMatcher interface {
	RegistMatcher(matcher ...ExpressionMatcher)
	GenerateSQL(item SqlState, sqlTpl string, input DBParam) (sql string, err error)
}

type DefaultTemplateMatcher struct {
	matcherMap ExpressionMatcherMap
	exprCache  *sync.Map
}

// 新建一个默认模板匹配器
func NewDefaultTemplateMatcher(matchers ...ExpressionMatcher) TemplateMatcher {

	conn := &DefaultTemplateMatcher{
		matcherMap: NewExpressionMatcherMap(),
		exprCache:  &sync.Map{},
	}
	conn.RegistMatcher(matchers...)
	return conn
}

func (conn *DefaultTemplateMatcher) RegistMatcher(matchers ...ExpressionMatcher) {
	conn.matcherMap.Regist(matchers...)
}

func (conn *DefaultTemplateMatcher) GenerateSQL(state SqlState, sqlTpl string, input DBParam) (sql string, err error) {

	matcherMap := conn.matcherMap
	word := matcherMap.BuildFullRegexp()

	var outerrs []MissError

	//@变量, 将数据放入params中
	sql = word.ReplaceAllStringFunc(sqlTpl, func(expr string) (repExpr string) {
		var (
			valuer  ExpressionValuer
			matcher ExpressionMatcher

			ok bool
		)

		exprItem, ok := conn.exprCache.Load(expr)
		if ok {
			matcher = exprItem.(ExpressionMatcher)
			valuer, _ = matcher.MatchString(expr)
		} else {
			matcherMap.Each(func(name string, exprMatcher ExpressionMatcher) bool {
				valuer, ok = exprMatcher.MatchString(expr)
				matcher = exprMatcher
				return !ok
			})

			if valuer == nil {
				return expr
			}
			if state.CanCache() {
				conn.exprCache.Store(expr, matcher)
			}
		}

		symbol, ok := matcher.LoadSymbol(valuer.GetSymbol())
		if !ok {
			return expr
		}
		repExpr, err := symbol.Callback(state, valuer, input)
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
