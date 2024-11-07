package xdb

import (
	"bytes"
	"regexp"
	"strings"
	"sync"

	"github.com/emirpasic/gods/v2/maps/treemap"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/golibs/xtypes"
)

// 表达式解析选项
type TemplateOptions struct {
	UseExprCache bool
}

type TemplateOption func(*TemplateOptions)

// 使用解析缓存
func WithExprCache(use bool) TemplateOption {
	return func(o *TemplateOptions) {
		o.UseExprCache = use
	}
}

type MatcherOptions struct {
	BuildCallback ExpressionBuildCallback
}
type MatcherOption func(*MatcherOptions)

// WithBuildCallback 制定matcher的表达式生成回调
func WithBuildCallback(callback ExpressionBuildCallback) MatcherOption {
	return func(mo *MatcherOptions) {
		mo.BuildCallback = callback
	}
}

// ExprCacheItem 表达式缓存
type ExprCacheItem struct {
	Matcher ExpressionMatcher
	Valuer  ExpressionValuer
}

// TemplateMatcher 模板匹配器
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
	word := matcherMap.GetMatcherRegexp()

	var outerrs []MissError

	//@变量, 将数据放入params中
	sql = word.ReplaceAllStringFunc(sqlTpl, func(expr string) (repExpr string) {
		var (
			matcher ExpressionMatcher
			valuer  ExpressionValuer
			ok      bool
		)

		exprItem, ok := conn.exprCache.Load(expr)
		if ok {
			expritem := exprItem.(*ExprCacheItem)
			matcher = expritem.Matcher
			valuer = expritem.Valuer
		} else {
			matcher = matcherMap.Find(func(exprMatcher ExpressionMatcher) bool {
				valuer, ok = exprMatcher.MatchString(expr)
				return ok
			})
			if valuer == nil {
				return expr
			}
			if state.UseExprCache() {
				conn.exprCache.Store(expr, &ExprCacheItem{
					Matcher: matcher,
					Valuer:  valuer,
				})
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
		return repExpr

	})
	if len(outerrs) > 0 {
		return sql, NewMissListError(outerrs...)
	}
	return sql, nil
}

type DefaultExpressionMatcherMapImpl struct {
	mutex        *sync.Mutex
	matcherCache *treemap.Map[int, ExpressionMatcher]
	sortVal      map[string]int
	regexp       *regexp.Regexp
}

func NewExpressionMatcherMap() ExpressionMatcherMap {
	return &DefaultExpressionMatcherMapImpl{
		mutex:        &sync.Mutex{},
		matcherCache: treemap.New[int, ExpressionMatcher](),
		sortVal:      map[string]int{},
	}
}

func (m *DefaultExpressionMatcherMapImpl) Regist(matchers ...ExpressionMatcher) {
	if len(matchers) <= 0 {
		return
	}
	if global.IsRunning() {
		return
	}

	for i := range matchers {
		matcher := matchers[i]
		if matcher == nil {
			continue
		}
		idx, ok := m.sortVal[matcher.Name()]
		if !ok {
			idx = len(m.sortVal)
			m.sortVal[matcher.Name()] = idx
		}

		m.matcherCache.Put(idx, matcher)
	}
}
func (m *DefaultExpressionMatcherMapImpl) Load(name string) (ExpressionMatcher, bool) {
	idx := m.sortVal[name]
	tmp, ok := m.matcherCache.Get(idx)
	if !ok {
		return nil, ok
	}
	return tmp, ok
}
func (m *DefaultExpressionMatcherMapImpl) Find(call func(matcher ExpressionMatcher) bool) ExpressionMatcher {
	_, matcher := m.matcherCache.Find(func(key int, value ExpressionMatcher) bool {
		return call(value)
	})
	return matcher
}
func (m *DefaultExpressionMatcherMapImpl) Delete(name string) {
	if global.IsRunning() {
		return
	}
	idx := m.sortVal[name]
	m.matcherCache.Remove(idx)
}

func (m *DefaultExpressionMatcherMapImpl) Clone() ExpressionMatcherMap {
	clone := NewExpressionMatcherMap()

	m.matcherCache.Each(func(idx int, matcher ExpressionMatcher) {
		clone.Regist(matcher)
	})

	return clone
}

func (m *DefaultExpressionMatcherMapImpl) GetMatcherRegexp() *regexp.Regexp {
	if m.regexp != nil {
		return m.regexp
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.regexp != nil {
		return m.regexp
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
	m.regexp = pattern
	return pattern
}
