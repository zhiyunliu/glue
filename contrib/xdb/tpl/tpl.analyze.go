package tpl

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/zhiyunliu/golibs/xsecurity/md5"
)

type SymbolCallback func(DBParam, string, *ReplaceItem) string
type Symbols map[string]SymbolCallback
type Placeholder interface {
	Get(propName string) (argName string, phName string)
	Clone() Placeholder
}

var tplcache sync.Map

// AnalyzeTPLFromCache 从缓存中获取已解析的SQL语句
// @表达式，替换为参数化字符如: :1,:2,:3
// $表达式，检查值，值为空时返加"",否则直接替换字符
// &条件表达式，检查值，值为空时返加"",否则返回: and name=value
// |条件表达式，检查值，值为空时返回"", 否则返回: or name=value
func AnalyzeTPLFromCache(template SQLTemplate, tpl string, input map[string]interface{}, ph Placeholder) (sql string, values []sql.NamedArg) {
	hashVal := md5.Str(template.Name() + tpl)
	tplval, ok := tplcache.Load(hashVal)
	if !ok {
		sql, rpsitem := template.AnalyzeTPL(tpl, input, ph)
		temp := &cacheItem{
			sql:         sql,
			names:       rpsitem.Names,
			SQLTemplate: template,
			ph:          ph.Clone(),
		}
		values = rpsitem.Values

		temp.nameCache = map[string]string{}
		for k := range rpsitem.NameCache {
			temp.nameCache[k] = rpsitem.NameCache[k]
		}

		sql, hasReplace := handleRelaceSymbols(sql, input, ph)
		temp.hasReplace = hasReplace

		sql, vals, hasAnd := template.HandleAndSymbols(sql, rpsitem, input)
		temp.hasDynamicAnd = hasAnd
		values = append(values, vals...)

		sql, vals, hasOr := template.HandleOrSymbols(sql, rpsitem, input)
		temp.hasDynamicOr = hasOr
		values = append(values, vals...)

		tplcache.Store(hashVal, temp)
		return sql, values
	}
	item := tplval.(*cacheItem)
	return item.build(input)
}

func DefaultAnalyze(symbols Symbols, tpl string, input map[string]interface{}, placeholder Placeholder) (string, *ReplaceItem) {
	word, _ := regexp.Compile(ParamPattern)
	item := &ReplaceItem{
		NameCache:   map[string]string{},
		Placeholder: placeholder,
	}
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
		return callback(input, fullKey, item)
	})

	return sql, item
}

func IsNil(input interface{}) bool {
	if input == nil {
		return true
	}
	if fmt.Sprintf("%v", input) == "" {
		return true
	}
	rv := reflect.ValueOf(input)
	if rv.Kind() == reflect.Ptr {
		return rv.IsNil()
	}
	return false
}

func GetPropName(fullKey string) (propName string) {
	propName = fullKey
	if strings.Index(fullKey, ".") > 0 {
		propName = strings.Split(fullKey, ".")[1]
	}
	return propName
}
