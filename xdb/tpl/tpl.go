package tpl

import (
	"fmt"
	"strings"
)

var (
	tpls map[string]SQLTemplate
)

//Template 模板上下文
type SQLTemplate interface {
	GetSQLContext(tpl string, input map[string]interface{}) (query string, args []interface{})
	GetSPContext(tpl string, input map[string]interface{}) (query string, args []interface{})
	Replace(sql string, args []interface{}) (r string)
}

func init() {
	tpls = make(map[string]SQLTemplate)

	Register("oracle", ATTPLContext{name: "oracle", prefix: ":"})
	Register("ora", ATTPLContext{name: "ora", prefix: ":"})
	Register("mysql", MTPLContext{name: "mysql", prefix: "?"})
	Register("sqlite", MTPLContext{name: "sqlite", prefix: "?"})
	Register("postgres", ATTPLContext{name: "postgres", prefix: "$"})
}
func Register(name string, tpl SQLTemplate) {
	if _, ok := tpls[name]; ok {
		panic("重复的注册:" + name)
	}
	tpls[name] = tpl
}

//GetDBTemplate 获取数据库上下文操作
func GetDBTemplate(name string) (SQLTemplate, error) {
	if v, ok := tpls[strings.ToLower(name)]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("不支持的数据库类型:%s", name)
}
