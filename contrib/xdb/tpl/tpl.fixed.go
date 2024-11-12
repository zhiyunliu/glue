package tpl

import (
	"fmt"

	"github.com/zhiyunliu/glue/xdb"
)

// FixedTemplate  模板
type FixedTemplate struct {
	name    string
	prefix  string
	matcher xdb.TemplateMatcher
}

type fixedPlaceHolder struct {
	template *FixedTemplate
}

func (ph *fixedPlaceHolder) Get(propName string) (argName, phName string) {
	phName = ph.template.prefix
	argName = propName
	return
}
func (ph *fixedPlaceHolder) BuildArgVal(argName string, val interface{}) interface{} {
	return val
}

func (ph *fixedPlaceHolder) NamedArg(propName string) (phName string) {
	phName = ph.template.prefix
	return
}

func (ph *fixedPlaceHolder) Clone() xdb.Placeholder {
	return &fixedPlaceHolder{
		template: ph.template,
	}
}

var _ xdb.SQLTemplate = &FixedTemplate{}

func NewFixed(name, prefix string, matcher xdb.TemplateMatcher) *FixedTemplate {
	if matcher == nil {
		panic(fmt.Errorf("NewFixed ,TemplateMatcher Can't be nil"))
	}
	return &FixedTemplate{
		name:    name,
		prefix:  prefix,
		matcher: matcher,
	}
}

func (template FixedTemplate) Name() string {
	return template.name
}

func (template *FixedTemplate) Placeholder() xdb.Placeholder {
	return &fixedPlaceHolder{template: template}
}

// GetSQLContext 获取查询串
func (template *FixedTemplate) GetSQLContext(sqlTpl string, input map[string]interface{}, opts ...xdb.TemplateOption) (query string, args []any, err error) {
	return AnalyzeTPLFromCache(template, sqlTpl, input, opts...)
}

func (template *FixedTemplate) RegistExpressionMatcher(matchers ...xdb.ExpressionMatcher) {
	template.matcher.RegistMatcher(matchers...)
}

func (template *FixedTemplate) HandleExpr(item xdb.SqlState, sqlTpl string, input xdb.DBParam) (sql string, err error) {
	return template.matcher.GenerateSQL(item, sqlTpl, input)
}
func (template *FixedTemplate) GetSqlState(tplOpts *xdb.TemplateOptions) xdb.SqlState {
	return xdb.NewSqlState(template.Placeholder(), tplOpts)
}
