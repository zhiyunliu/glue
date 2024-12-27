package sqlserver

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

// MssqlTemplate  模板
type MssqlTemplate struct {
	name          string
	prefix        string
	matcher       xdb.TemplateMatcher
	stmtProcessor xdb.StmtDbTypeProcessor
	sqlStatePool  *sync.Pool
}

type mssqlPlaceHolder struct {
	template *MssqlTemplate
}

func (ph *mssqlPlaceHolder) Get(propName string) (argName, phName string) {
	argName = fmt.Sprint(ph.template.prefix, propName)
	phName = ph.NamedArg(argName)
	return
}

func (ph *mssqlPlaceHolder) NamedArg(argName string) (phName string) {
	phName = "@" + argName
	return
}

func (ph *mssqlPlaceHolder) BuildArgVal(argName string, val any) any {
	if arg, ok := val.(sql.NamedArg); ok {
		return arg
	}
	return sql.NamedArg{Name: argName, Value: val}

}

func New(name, prefix string, matcher xdb.TemplateMatcher, stmtProcessor xdb.StmtDbTypeProcessor) xdb.SQLTemplate {

	if matcher == nil {
		panic(fmt.Errorf("New ,TemplateMatcher Can't be nil"))
	}
	template := &MssqlTemplate{
		name:          name,
		prefix:        prefix,
		matcher:       matcher,
		stmtProcessor: stmtProcessor,
	}

	template.sqlStatePool = &sync.Pool{
		New: func() interface{} {
			return NewSqlState(template.Placeholder())
		},
	}
	return template
}

func (template *MssqlTemplate) Name() string {
	return template.name
}

func (template *MssqlTemplate) Placeholder() xdb.Placeholder {
	return &mssqlPlaceHolder{template: template}
}

// GetSQLContext 获取查询串
func (template *MssqlTemplate) GetSQLContext(sqlTpl string, input map[string]any, opts ...xdb.TemplateOption) (query string, args []any, err error) {
	return tpl.AnalyzeTPLFromCache(template, sqlTpl, input, opts...)
}

func (template *MssqlTemplate) RegistExpressionMatcher(matchers ...xdb.ExpressionMatcher) {
	template.matcher.RegistMatcher(matchers...)
}

func (template *MssqlTemplate) HandleExpr(item xdb.SqlState, sqlTpl string, input xdb.DBParam) (sql string, err error) {
	return template.matcher.GenerateSQL(item, sqlTpl, input)
}

func (template *MssqlTemplate) GetSqlState(tplOpts *xdb.TemplateOptions) xdb.SqlState {
	sqlState := template.sqlStatePool.Get().(xdb.SqlState)
	sqlState.WithTemplateOptions(tplOpts)
	return sqlState
}

func (template *MssqlTemplate) ReleaseSqlState(state xdb.SqlState) {
	state.Reset()
	template.sqlStatePool.Put(state)
}

func (template *MssqlTemplate) StmtDbTypeWrap(param any, tagOpts xdb.TagOptions) any {
	return template.stmtProcessor.Process(param, tagOpts)
}
func (template *MssqlTemplate) RegistStmtDbTypeHandler(handler ...xdb.StmtDbTypeHandler) {
	template.stmtProcessor.RegistHandler(handler...)
}
