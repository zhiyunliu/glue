package tpl

import (
	"bytes"
	"regexp"
	"strings"
	"sync"

	"github.com/emirpasic/gods/v2/maps/treemap"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xsecurity/md5"
	"github.com/zhiyunliu/golibs/xtypes"
)

func initMatcher() {
	xdb.NewTemplateMatcher = NewDefaultTemplateMatcher
}

// ExpressionMatcher 默认表达式匹配器
type DefaultTemplateMatcher struct {
	matcherMap xdb.ExpressionMatcherMap
	exprCache  *sync.Map
}

// 新建一个默认模板匹配器
func NewDefaultTemplateMatcher(matchers ...xdb.ExpressionMatcher) xdb.TemplateMatcher {

	conn := &DefaultTemplateMatcher{
		matcherMap: NewExpressionMatcherMap(),
		exprCache:  &sync.Map{},
	}
	conn.RegistMatcher(matchers...)
	return conn
}

func (conn *DefaultTemplateMatcher) RegistMatcher(matchers ...xdb.ExpressionMatcher) {
	conn.matcherMap.Regist(matchers...)
}

func (conn *DefaultTemplateMatcher) GenerateSQL(state xdb.SqlState, sqlTpl string, input xdb.DBParam) (sql string, err error) {

	tplHash := md5.Str(sqlTpl)

	if tmp, ok := conn.exprCache.Load(tplHash); ok {
		tplCache := tmp.(xdb.ExpressionCache)
		return tplCache.Build(state, input)
	}

	matcherMap := conn.matcherMap
	word := matcherMap.GetMatcherRegexp()

	var outerrs []xdb.MissError

	//@变量, 将数据放入params中
	sql = word.ReplaceAllStringFunc(sqlTpl, func(expr string) (resultExpr string) {
		var (
			valuer xdb.ExpressionValuer
			ok     bool
		)

		exprItem, ok := conn.exprCache.Load(expr)
		if ok {
			valuer = exprItem.(xdb.ExpressionValuer)
		} else {
			matcherMap.Find(func(exprMatcher xdb.ExpressionMatcher) bool {
				valuer, ok = exprMatcher.MatchString(expr)
				return ok
			})
			if valuer == nil {
				return expr
			}
			if state.UseExprCache() {
				conn.exprCache.Store(expr, valuer)
			}
		}

		resultExpr, err := valuer.Build(state, input)
		if err != nil {
			outerrs = append(outerrs, err)
			return expr
		}
		return resultExpr

	})
	if len(outerrs) > 0 {
		return sql, xdb.NewMissListError(outerrs...)
	}

	if state.CanCache() {
		tplCache := state.BuildCache(sql)
		conn.exprCache.Store(tplHash, tplCache)
	}

	return sql, nil
}

// DefaultExpressionMatcherMapImpl 默认表达式匹配器实现
type DefaultExpressionMatcherMapImpl struct {
	mutex        *sync.Mutex
	matcherCache *treemap.Map[int, xdb.ExpressionMatcher]
	sortVal      map[string]int
	regexp       *regexp.Regexp
}

func NewExpressionMatcherMap() xdb.ExpressionMatcherMap {
	return &DefaultExpressionMatcherMapImpl{
		mutex:        &sync.Mutex{},
		matcherCache: treemap.New[int, xdb.ExpressionMatcher](),
		sortVal:      map[string]int{},
	}
}

func (m *DefaultExpressionMatcherMapImpl) Regist(matchers ...xdb.ExpressionMatcher) {
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
func (m *DefaultExpressionMatcherMapImpl) Load(name string) (xdb.ExpressionMatcher, bool) {
	idx := m.sortVal[name]
	tmp, ok := m.matcherCache.Get(idx)
	if !ok {
		return nil, ok
	}
	return tmp, ok
}
func (m *DefaultExpressionMatcherMapImpl) Find(call func(matcher xdb.ExpressionMatcher) bool) xdb.ExpressionMatcher {
	_, matcher := m.matcherCache.Find(func(key int, value xdb.ExpressionMatcher) bool {
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

func (m *DefaultExpressionMatcherMapImpl) Clone() xdb.ExpressionMatcherMap {
	clone := NewExpressionMatcherMap()

	m.matcherCache.Each(func(idx int, matcher xdb.ExpressionMatcher) {
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
