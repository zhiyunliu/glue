package tpl

import (
	"regexp"
	"strings"
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

func DefaultAnalyze(symbols Symbols, tpl string, input map[string]interface{}, placeholder Placeholder) (string, *ReplaceItem, error) {
	word, _ := regexp.Compile(TotalPattern)
	item := &ReplaceItem{
		NameCache:   map[string]string{},
		Placeholder: placeholder,
	}

	var outerrs []xdb.MissParamError

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

		callback, ok := symbols[symbol]
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
		return sql, item, xdb.NewMissParamsError(outerrs...)
	}

	return sql, item, nil
}

func GetPropName(fullKey string) (propName string) {
	propName = fullKey
	if strings.Index(fullKey, ".") > 0 {
		propName = strings.Split(fullKey, ".")[1]
	}
	return propName
}
