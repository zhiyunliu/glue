package sqlserver

import (
	"database/sql"
	"fmt"

	"github.com/zhiyunliu/glue/contrib/xdb/tpl"
	"github.com/zhiyunliu/glue/xdb"
)

// MssqlContext  模板
type MssqlContext struct {
	name    string
	prefix  string
	matcher xdb.TemplateMatcher
}

type mssqlPlaceHolder struct {
	ctx *MssqlContext
}

func (ph *mssqlPlaceHolder) Get(propName string) (argName, phName string) {
	argName = fmt.Sprint(ph.ctx.prefix, propName)
	phName = "@" + argName
	return
}

func (ph *mssqlPlaceHolder) NamedArg(argName string) (phName string) {
	phName = "@" + argName
	return
}

func (ph *mssqlPlaceHolder) BuildArgVal(argName string, val interface{}) interface{} {
	if arg, ok := val.(sql.NamedArg); ok {
		return arg
	}
	return sql.NamedArg{Name: argName, Value: val}

}

func (ph *mssqlPlaceHolder) Clone() xdb.Placeholder {
	return &mssqlPlaceHolder{
		ctx: ph.ctx,
	}
}

func New(name, prefix string, matcher xdb.TemplateMatcher) xdb.SQLTemplate {

	if matcher == nil {
		panic(fmt.Errorf("New ,TemplateMatcher Can't be nil"))
	}
	return &MssqlContext{
		name:    name,
		prefix:  prefix,
		matcher: matcher,
	}
}

func (ctx *MssqlContext) Name() string {
	return ctx.name
}

func (ctx *MssqlContext) Placeholder() xdb.Placeholder {
	return &mssqlPlaceHolder{ctx: ctx}
}

// GetSQLContext 获取查询串
func (template *MssqlContext) GetSQLContext(sqlTpl string, input map[string]any, opts ...xdb.TemplateOption) (query string, args []any, err error) {
	return tpl.AnalyzeTPLFromCache(template, sqlTpl, input, opts...)
}

func (template *MssqlContext) RegistExpressionMatcher(matchers ...xdb.ExpressionMatcher) {
	template.matcher.RegistMatcher(matchers...)
}

func (template *MssqlContext) HandleExpr(item xdb.SqlState, sqlTpl string, input xdb.DBParam) (sql string, err error) {
	return template.matcher.GenerateSQL(item, sqlTpl, input)
}

func (template *MssqlContext) GetSqlState(tplOpts *xdb.TemplateOptions) xdb.SqlState {
	return xdb.NewDefaultSqlState(template.Placeholder(), tplOpts)
}
