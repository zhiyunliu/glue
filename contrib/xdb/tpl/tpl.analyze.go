package tpl

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/zhiyunliu/golibs/xsecurity/md5"
)

type SymbolCallback func(map[string]interface{}, string, *ReplaceItem) string
type Symbols map[string]SymbolCallback
type Placeholder func() string
type cacheItem struct {
	sql        string
	hasReplace bool
	names      []string
}

type ReplaceItem struct {
	Names       []string
	Values      []interface{}
	NameCache   map[string]string
	Placeholder Placeholder
}

var tplcache sync.Map

var defaultSymbols Symbols

func init() {
	defaultSymbols = make(Symbols)
	defaultSymbols["@"] = func(input map[string]interface{}, fullKey string, item *ReplaceItem) string {
		propName := GetPropName(fullKey)
		value := input[propName]
		if !IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
		} else {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, "")
		}
		return item.Placeholder()
	}

	defaultSymbols["&"] = func(input map[string]interface{}, fullKey string, item *ReplaceItem) string {
		propName := GetPropName(fullKey)
		value := input[propName]
		if !IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return fmt.Sprintf(" and %s=%s", fullKey, item.Placeholder())
		}
		return ""
	}
	defaultSymbols["|"] = func(input map[string]interface{}, fullKey string, item *ReplaceItem) string {
		propName := GetPropName(fullKey)
		value := input[propName]
		if !IsNil(value) {
			item.Names = append(item.Names, propName)
			item.Values = append(item.Values, value)
			return fmt.Sprintf(" or %s=%s", fullKey, item.Placeholder())
		}
		return ""
	}

}

func (item cacheItem) build(input map[string]interface{}) (sql string, values []interface{}) {
	values = make([]interface{}, len(item.names))
	for i := range item.names {
		values[i] = input[item.names[i]]
	}
	sql = item.sql
	if item.hasReplace {
		sql, _ = handleRelaceSymbols(item.sql, input)
	}
	return sql, values
}

//AnalyzeTPLFromCache 从缓存中获取已解析的SQL语句
//@表达式，替换为参数化字符如: :1,:2,:3
//$表达式，检查值，值为空时返加"",否则直接替换字符
//&条件表达式，检查值，值为空时返加"",否则返回: and name=value
//|条件表达式，检查值，值为空时返回"", 否则返回: or name=value
func AnalyzeTPLFromCache(template SQLTemplate, tpl string, input map[string]interface{}) (sql string, values []interface{}) {
	hashVal := md5.Str(template.Name() + tpl)
	tplval, ok := tplcache.Load(hashVal)
	if !ok {
		sql, names, values := template.AnalyzeTPL(tpl, input)
		temp := &cacheItem{
			sql:   sql,
			names: names,
		}
		sql, hasReplace := handleRelaceSymbols(sql, input)
		temp.hasReplace = hasReplace
		tplcache.Store(hashVal, temp)
		return sql, values
	}
	item := tplval.(*cacheItem)
	return item.build(input)
}

func DefaultAnalyze(symbols Symbols, tpl string, input map[string]interface{}, placeholder func() string) (string, []string, []interface{}) {
	word, _ := regexp.Compile(`[@|#|&|\|]\{\w+[\.]?\w+\}`)
	item := &ReplaceItem{
		NameCache:   map[string]string{},
		Placeholder: placeholder,
	}
	//@变量, 将数据放入params中
	sql := word.ReplaceAllStringFunc(tpl, func(s string) string {
		/*
			@{aaaa}
			${bbb}
			${c.cc}
			&{ddd}
			~{asdfasdf}
			&{t.asdfasdf}
			#{aaaa.b}
			|{aaaa.b}
		*/
		symbol := s[:1]
		fullKey := s[2 : len(s)-1]

		callback, ok := symbols[symbol]
		if !ok {
			return s
		}
		return callback(input, fullKey, item)
	})

	return sql, item.Names, item.Values
}

func IsNil(input interface{}) bool {
	return input == nil || fmt.Sprintf("%v", input) == ""
}

func GetPropName(fullKey string) (propName string) {
	propName = fullKey
	if strings.Index(fullKey, ".") > 0 {
		propName = strings.Split(fullKey, ".")[1]
	}
	return propName
}

//处理替换符合
func handleRelaceSymbols(tpl string, input map[string]interface{}) (string, bool) {
	word, _ := regexp.Compile(`\$\{\w+[\.]?\w+\}`)
	item := &ReplaceItem{
		NameCache: map[string]string{},
	}
	hasReplace := false
	sql := word.ReplaceAllStringFunc(tpl, func(s string) string {
		hasReplace = true
		fullKey := s[2 : len(s)-1]
		return replaceSymbols(input, fullKey, item)
	})

	return sql, hasReplace
}

func replaceSymbols(input map[string]interface{}, fullKey string, item *ReplaceItem) string {
	propName := GetPropName(fullKey)
	value := input[propName]
	if !IsNil(value) {
		return fmt.Sprintf("%v", value)
	}
	return ""
}
