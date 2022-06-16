package tpl

import (
	"fmt"
	"strings"

	"github.com/zhiyunliu/glue/log"
)

var (
	tpls map[string]SQLTemplate
)

//Template 模板上下文
type SQLTemplate interface {
	Name() string
	Placeholder() Placeholder
	GetSQLContext(tpl string, input map[string]interface{}) (query string, args []interface{})
	AnalyzeTPL(tpl string, input map[string]interface{}) (sql string, names []string, values []interface{})
}

func init() {
	tpls = make(map[string]SQLTemplate)
}
func Register(tpl SQLTemplate) {
	if _, ok := tpls[tpl.Name()]; ok {
		log.Warnf("%s 注册北覆盖", tpl.Name())
	}
	tpls[tpl.Name()] = tpl
}

//GetDBTemplate 获取数据库上下文操作
func GetDBTemplate(name string) (SQLTemplate, error) {
	if v, ok := tpls[strings.ToLower(name)]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("不支持的数据库类型:%s", name)
}
