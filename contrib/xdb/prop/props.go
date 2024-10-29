package prop

import (
	"strings"
	"sync"

	"github.com/emirpasic/gods/v2/maps/treemap"
	"github.com/zhiyunliu/glue/xdb"
)

var (
	propMatcher  *treemap.Map[int, xdb.PropMatcher] = treemap.New[int, xdb.PropMatcher]()
	fullkeyCache sync.Map
)

// type PropCallback func()
// type MatchCallback func(string) (fullField, propName, oper string)
type PropertyCallback func(item *PropItem, symbol string, param xdb.DBParam, argName string) (part string, err xdb.MissError)

type PropItem struct {
	FullField        string
	PropName         string
	Oper             string
	PropertyCallback PropertyCallback
}

func (m *PropItem) GetPropName() string {
	return m.PropName
}

func (m *PropItem) GetFullfield() string {
	return m.FullField
}

func (m *PropItem) GetOper() string {
	return m.Oper
}

func (m *PropItem) Build(symbol string, param xdb.DBParam, argName string) (part string, err xdb.MissError) {
	if m.PropertyCallback == nil {
		return
	}
	return m.PropertyCallback(m, symbol, param, argName)
}

func RegistMatcher(matcher xdb.PropMatcher) {
	propMatcher.Put(matcher.Priority(), matcher)
}

// 默认的获取属性值的匹配方法
func DefaultGetPropMatchValuer(fullKey string, opts *xdb.PropOptions) (valuer xdb.PropValuer) {
	fullKey = strings.TrimSpace(fullKey)

	if opts.UseCache {
		if tmp, ok := fullkeyCache.Load(fullKey); ok {
			valuer = tmp.(xdb.PropValuer)
			return
		}
	}
	var (
		ok bool
	)
	propMatcher.Find(func(key int, value xdb.PropMatcher) bool {
		valuer, ok = value.MatchString(fullKey)
		return ok
	})

	if opts.UseCache {
		fullkeyCache.Store(fullKey, valuer)
	}
	return

	// fullField = propName
	// idx := strings.Index(propName, " ")
	// if idx < 0 {
	// 	// <tbl.field,<=tbl.field,>tbl.field,>=tbl.field
	// 	switch {
	// 	case strings.HasPrefix(fullField, "<="): //<=tbl.field,<=field
	// 		propName = strings.TrimPrefix(fullField, "<=")
	// 		fullField = propName
	// 		oper = "<="
	// 	case strings.HasPrefix(fullField, "<"):
	// 		propName = strings.TrimPrefix(fullField, "<")
	// 		fullField = propName
	// 		oper = "<"
	// 	case strings.HasPrefix(fullField, ">="):
	// 		propName = strings.TrimPrefix(fullField, ">=")
	// 		fullField = propName
	// 		oper = ">="
	// 	case strings.HasPrefix(fullField, ">"):
	// 		propName = strings.TrimPrefix(fullField, ">")
	// 		fullField = propName
	// 		oper = ">"
	// 	default:
	// 		oper = "="
	// 	}

	// 	if strings.Index(propName, ".") > 0 {
	// 		propName = strings.Split(propName, ".")[1]
	// 	}
	// 	return fullField, propName, oper
	// }

	// parties := strings.Split(propName, " ")

	// tmpfield := parties[len(parties)-1]
	// fullField = strings.Trim(tmpfield, "%")
	// propName, oper = procLike(tmpfield, parties[0])

	// if strings.Index(propName, ".") > 0 {
	// 	propName = strings.Split(propName, ".")[1]
	// }
	// return fullField, propName, oper
}

// func procLike(filed, orgOper string) (propName, oper string) {
// 	orgOper = strings.TrimSpace(orgOper)
// 	filed = strings.TrimSpace(filed)

// 	if !strings.EqualFold(orgOper, "like") {
// 		oper = orgOper
// 		propName = filed
// 		return
// 	}

// 	var (
// 		prefix string = ""
// 		suffix string = ""
// 	)

// 	if strings.HasPrefix(filed, "%") {
// 		prefix = "%"
// 	}
// 	if strings.HasSuffix(filed, "%") {
// 		suffix = "%"
// 	}
// 	oper = prefix + orgOper + suffix
// 	propName = strings.Trim(filed, "%")
// 	return
// }

func getPropertyName(fullkey string) string {
	idx := strings.Index(fullkey, ".")
	if idx < 0 {
		return fullkey
	}
	return fullkey[idx+1:]
}
