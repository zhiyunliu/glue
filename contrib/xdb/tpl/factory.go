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
	Name() string
	Placeholder() Placeholder
	GetSQLContext(tpl string, input map[string]interface{}) (query string, args []interface{})
	analyzeTPL(tpl string, input map[string]interface{}) (sql string, params []interface{}, names []string)
}

func init() {
	tpls = make(map[string]SQLTemplate)

	Register("mysql", &FixedContext{name: "mysql", prefix: "?"})
	Register("sqlite", &FixedContext{name: "sqlite", prefix: "?"})
	Register("oracle", &SeqContext{name: "oracle", prefix: ":"})
	Register("postgres", &SeqContext{name: "postgres", prefix: "$"})
	Register("sqlserver", &MssqlContext{name: "sqlserver", prefix: "@p"})
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
