package tpl

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/zhiyunliu/golibs/xsecurity/md5"
)

var tplcache sync.Map

type Placeholder func() string
type cacheItem struct {
	sql   string
	names []string
}

func (item cacheItem) build(input map[string]interface{}) (params []interface{}) {
	params = make([]interface{}, len(item.names))
	for i := range item.names {
		params[i] = input[item.names[i]]
	}
	return
}

func isNil(input interface{}) bool {
	return input == nil || fmt.Sprintf("%v", input) == ""
}

//AnalyzeTPLFromCache 从缓存中获取已解析的SQL语句
func AnalyzeTPLFromCache(template SQLTemplate, tpl string, input map[string]interface{}) (sql string, params []interface{}) {
	hashVal := md5.Str(template.Name() + tpl)
	tplval, ok := tplcache.Load(hashVal)
	if !ok {
		sql, params, names := AnalyzeTPL(template, tpl, input)
		tplval = &cacheItem{
			sql:   sql,
			names: names,
		}
		tplcache.Store(hashVal, tplval)
		return sql, params
	}

	item := tplval.(*cacheItem)

	return item.sql, item.build(input)
}

//AnalyzeTPL 解析模板内容，并返回解析后的SQL语句，入输入参数
//@表达式，替换为参数化字符如: :1,:2,:3
//#表达式，替换为指定值，值为空时返回NULL
//~表达式，检查值，值为空时返加"",否则返回: , name=value
//&条件表达式，检查值，值为空时返加"",否则返回: and name=value
//|条件表达式，检查值，值为空时返回"", 否则返回: or name=value
func AnalyzeTPL(template SQLTemplate, tpl string, input map[string]interface{}) (sql string, params []interface{}, names []string) {
	return template.analyzeTPL(tpl, input)
}

func defaultAnalyze(tpl string, input map[string]interface{}, placeholder func() string) (sql string, params []interface{}, names []string) {
	params = make([]interface{}, 0)
	names = make([]string, 0)

	word, _ := regexp.Compile(`[\\]?[@|#|&|~|\||!|\$|\?]\w?[\.]?\w+`)

	//@变量, 将数据放入params中
	sql = word.ReplaceAllStringFunc(tpl, func(s string) string {
		fullKey, key, name := s[1:], s[1:], s[1:]
		if strings.Index(fullKey, ".") > 0 {
			name = strings.Split(fullKey, ".")[1]
		}
		pre := s[:1]
		value := input[name]
		switch pre {
		case "@":
			if !isNil(value) {
				names = append(names, key)
				params = append(params, value)
			} else {
				names = append(names, key)
				params = append(params, "")
			}
			return placeholder()
		case "$":
			if !isNil(value) {
				return fmt.Sprintf("%v", value)
			}
			return ""
		case "&":
			if !isNil(value) {
				names = append(names, key)
				params = append(params, value)
				return fmt.Sprintf("and %s=%s", key, placeholder())
			}
			return ""
		case "|":
			if !isNil(value) {
				names = append(names, key)
				params = append(params, value)
				return fmt.Sprintf("or %s=%s", key, placeholder())
			}
			return ""
		default:
			return s
		}
	})

	// word2, _ := regexp.Compile(`[\\][@|#|&|~|\||!|\$|\?|>|<]`)
	// //@变量, 将数据放入params中
	// sql = word2.ReplaceAllStringFunc(sql, func(s string) string {
	// 	return s[1:]
	// })
	return
}
