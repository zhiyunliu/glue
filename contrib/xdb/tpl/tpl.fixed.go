package tpl

import (
	"fmt"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

// FixedTemplate  模板
type FixedTemplate struct {
	name          string
	prefix        string
	matcher       xdb.TemplateMatcher
	stmtProcessor xdb.StmtDbTypeProcessor
	sqlStatePool  *sync.Pool
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

var _ xdb.SQLTemplate = &FixedTemplate{}

func NewFixed(name, prefix string, matcher xdb.TemplateMatcher, stmtProcessor xdb.StmtDbTypeProcessor) *FixedTemplate {
	if matcher == nil {
		panic(fmt.Errorf("NewFixed ,TemplateMatcher Can't be nil"))
	}
	template := &FixedTemplate{
		name:          name,
		prefix:        prefix,
		matcher:       matcher,
		stmtProcessor: stmtProcessor,
	}
	template.sqlStatePool = &sync.Pool{
		New: func() interface{} {
			return xdb.NewSqlState(template.Placeholder())
		},
	}
	return template
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
	sqlState := template.sqlStatePool.Get().(xdb.SqlState)
	sqlState.WithTemplateOptions(tplOpts)
	return sqlState
}

func (template *FixedTemplate) ReleaseSqlState(state xdb.SqlState) {
	state.Reset()
	template.sqlStatePool.Put(state)
}

func (template *FixedTemplate) StmtDbTypeWrap(param any, tagOpts xdb.TagOptions) any {
	return template.stmtProcessor.Process(param, tagOpts)
}
func (template *FixedTemplate) RegistStmtDbTypeHandler(handler ...xdb.StmtDbTypeHandler) {
	template.stmtProcessor.RegistHandler(handler...)
}
