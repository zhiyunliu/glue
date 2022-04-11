package tpl

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

//MssqlContext  模板
type MssqlContext struct {
	name   string
	prefix string
}

func (ctx *MssqlContext) Name() string {
	return ctx.name
}

//GetSQLContext 获取查询串
func (ctx *MssqlContext) GetSQLContext(tpl string, input map[string]interface{}) (query string, args []interface{}) {
	return AnalyzeTPLFromCache(ctx, tpl, input)
}

func (ctx *MssqlContext) Placeholder() Placeholder {
	index := 0
	f := func() string {
		index++
		return fmt.Sprint(ctx.prefix, index)
	}
	return f
}

func (ctx *MssqlContext) analyzeTPL(tpl string, input map[string]interface{}) (sql string, params []interface{}, names []string) {
	params = make([]interface{}, 0)
	names = make([]string, 0)
	placeholder := ctx.Placeholder()
	word, _ := regexp.Compile(`[\\]?[#|&|~|\||!|\$|\?]\w?[\.]?\w+`)

	cacheParam := map[string]string{}

	//@变量, 将数据放入params中
	sql = word.ReplaceAllStringFunc(tpl, func(s string) string {
		fullKey, key, name := s[1:], s[1:], s[1:]
		if strings.Index(fullKey, ".") > 0 {
			name = strings.Split(fullKey, ".")[1]
		}
		pre := s[:1]
		value := input[name]
		switch pre {
		case "#":
			if ph, ok := cacheParam[key]; ok {
				return ph
			}
			if !isNil(value) {
				names = append(names, key)
				params = append(params, value)
			} else {
				names = append(names, key)
				params = append(params, "")
			}
			cacheParam[key] = getPlaceHolder(value, placeholder)
			return cacheParam[key]
		case "$":
			if !isNil(value) {
				return fmt.Sprintf("%v", value)
			}
			return ""
		case "&":
			if ph, ok := cacheParam[key]; ok {
				return ph
			}

			if !isNil(value) {
				names = append(names, key)
				params = append(params, value)
				cacheParam[key] = getPlaceHolder(value, placeholder)
				return fmt.Sprintf(" and %s=%s", key, cacheParam[key])
			}
			return ""
		case "|":
			if ph, ok := cacheParam[key]; ok {
				return ph
			}
			if !isNil(value) {
				names = append(names, key)
				params = append(params, value)
				cacheParam[key] = getPlaceHolder(value, placeholder)
				return fmt.Sprintf(" or %s=%s", key, cacheParam[key])
			}
			return ""
		default:
			return s
		}
	})
	return
}

func getPlaceHolder(value interface{}, placeholder Placeholder) string {
	if arg, ok := value.(sql.NamedArg); ok {
		return "@" + arg.Name
	}
	return placeholder()
}
