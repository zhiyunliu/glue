package xdb

import (
	"fmt"
	"strings"
	"sync"
)

var (
	tpls sync.Map
)

type SqlScene struct {
	Names             []string
	Values            []interface{}
	NameCache         map[string]string
	Placeholder       Placeholder
	PropOpts          *ExpressionOptions
	HasDynamicAnd     bool
	HasDynamicOr      bool
	HasDynamicReplace bool
}

func (p *SqlScene) Clone() *SqlScene {
	return &SqlScene{
		NameCache:   p.NameCache,
		Placeholder: p.Placeholder,
	}
}

// 是否缓存
func (p *SqlScene) CanCache() bool {
	return !(p.HasDynamicAnd || p.HasDynamicOr)
}

// Template 模板上下文
type SQLTemplate interface {
	Name() string
	Placeholder() Placeholder
	GetSQLContext(tpl string, input map[string]any) (query string, args []any, err error)
	RegistExpressionMatcher(matchers ...ExpressionMatcher)
	GenerateSQL(item *SqlScene, sqlTpl string, param DBParam) (sql string, err error)
}

func RegistTemplate(tpl SQLTemplate) (err error) {
	name := strings.ToLower(tpl.Name())
	tpls.Store(name, tpl)
	return
}

// GetDBTemplate 获取数据库上下文操作
func GetTemplate(name string) (SQLTemplate, error) {
	if v, ok := tpls.Load(strings.ToLower(name)); ok {
		return v.(SQLTemplate), nil
	}
	return nil, fmt.Errorf("不支持的数据库类型:%s", name)
}

// RegistExpressionMatcher 给数据库注册语法解析
func RegistExpressionMatcher(proto string, matcher ExpressionMatcher) (err error) {
	tmpl, err := GetTemplate(proto)
	if err != nil {
		return err
	}
	tmpl.RegistExpressionMatcher(matcher)
	return
}
