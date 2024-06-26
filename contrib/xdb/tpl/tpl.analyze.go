package tpl

import (
	"regexp"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/golibs/xsecurity/md5"
)

var tplcache sync.Map

// AnalyzeTPLFromCache 从缓存中获取已解析的SQL语句
// @表达式，替换为参数化字符如: :1,:2,:3
// $表达式，检查值，值为空时返加"",否则直接替换字符
// &条件表达式，检查值，值为空时返加"",否则返回: and name=value
// |条件表达式，检查值，值为空时返回"", 否则返回: or name=value
func AnalyzeTPLFromCache(template SQLTemplate, tpl string, input map[string]interface{}, ph Placeholder) (sql string, values []any, err error) {
	hashVal := md5.Str(template.Name() + tpl)
	tplval, ok := tplcache.Load(hashVal)
	if !ok {
		sql, rpsitem, err := template.AnalyzeTPL(tpl, input, ph)
		if err != nil {
			return "", nil, err
		}

		values = rpsitem.Values
		if rpsitem.CanCache() {
			temp := &cacheItem{
				sql:         sql,
				names:       rpsitem.Names,
				SQLTemplate: template,
				ph:          ph.Clone(),
			}

			temp.nameCache = map[string]string{}
			for k := range rpsitem.NameCache {
				temp.nameCache[k] = rpsitem.NameCache[k]
			}
			sql, temp.hasReplace, err = handleRelaceSymbols(sql, input, ph)
			if err != nil {
				return sql, values, err
			}
			tplcache.Store(hashVal, temp)
		} else {
			sql, _, err = handleRelaceSymbols(sql, input, ph)
		}

		return sql, values, err
	}
	item := tplval.(*cacheItem)
	return item.build(input)
}

func DefaultAnalyze(symbols SymbolMap, tpl string, input map[string]interface{}, placeholder Placeholder) (string, *ReplaceItem, error) {
	word := GetPatternRegexp(symbols.GetPattern())
	item := &ReplaceItem{
		NameCache:   map[string]string{},
		Placeholder: placeholder,
	}

	var outerrs []xdb.MissError

	//@变量, 将数据放入params中
	sql := word.ReplaceAllStringFunc(tpl, func(s string) string {
		/*
			@{aaaa}
			@{t.aaaa}
			${cc}
			${c.cc}
			&{ddd}
			&{t.ddd}
			|{aaaa}
			|{t.aaaa}
		*/
		symbol := s[:1]
		fullKey := s[2 : len(s)-1]

		callback, ok := symbols.Load(symbol)
		if !ok {
			return s
		}
		tmpv, err := callback(input, fullKey, item)
		if err != nil {
			outerrs = append(outerrs, err)
		}
		return tmpv
	})
	if len(outerrs) > 0 {
		return sql, item, xdb.NewMissListError(outerrs...)
	}

	return sql, item, nil
}

// 获取模式匹配的正则表达式
func GetPatternRegexp(pattern string) *regexp.Regexp {
	tmpregex, ok := tplcache.Load(pattern)
	if ok {
		return tmpregex.(*regexp.Regexp)
	}

	act, _ := tplcache.LoadOrStore(pattern, regexp.MustCompile(pattern))
	return act.(*regexp.Regexp)
}
