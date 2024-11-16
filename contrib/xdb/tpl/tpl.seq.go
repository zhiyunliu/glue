package tpl

import (
	"fmt"
	"sync"

	"github.com/zhiyunliu/glue/xdb"
)

// SeqTemplate 参数化时使用@+参数名作为占位符的SQL数据库如:oracle,sql server
type SeqTemplate struct {
	name         string
	prefix       string
	matcher      xdb.TemplateMatcher
	sqlStatePool *sync.Pool
}

type seqPlaceHolder struct {
	template *SeqTemplate
	idx      int
}

func (ph *seqPlaceHolder) Get(propName string) (argName, phName string) {
	ph.idx++
	phName = fmt.Sprint(ph.template.prefix, ph.idx)
	argName = "p_" + propName
	return
}

func (ph *seqPlaceHolder) BuildArgVal(argName string, val interface{}) interface{} {
	return val
}

func (ph *seqPlaceHolder) NamedArg(propName string) (phName string) {
	ph.idx++
	phName = fmt.Sprint(ph.template.prefix, ph.idx)
	return
}

var _ xdb.SQLTemplate = &SeqTemplate{}

func NewSeq(name, prefix string, matcher xdb.TemplateMatcher) *SeqTemplate {
	if matcher == nil {
		panic(fmt.Errorf("NewFixed ,TemplateMatcher Can't be nil"))
	}
	template := &SeqTemplate{
		name:    name,
		prefix:  prefix,
		matcher: matcher,
	}

	template.sqlStatePool = &sync.Pool{
		New: func() interface{} {
			return xdb.NewSqlState(template.Placeholder())
		},
	}
	return template
}

func (template SeqTemplate) Name() string {
	return template.name
}

func (template *SeqTemplate) Placeholder() xdb.Placeholder {
	return &seqPlaceHolder{template: template, idx: 0}
}

// GetSQLContext 获取查询串
func (template *SeqTemplate) GetSQLContext(sqlTpl string, input map[string]interface{}, opts ...xdb.TemplateOption) (query string, args []any, err error) {
	return AnalyzeTPLFromCache(template, sqlTpl, input, opts...)
}

func (template *SeqTemplate) RegistExpressionMatcher(matchers ...xdb.ExpressionMatcher) {
	template.matcher.RegistMatcher(matchers...)
}

func (template *SeqTemplate) HandleExpr(item xdb.SqlState, sqlTpl string, input xdb.DBParam) (sql string, err error) {
	return template.matcher.GenerateSQL(item, sqlTpl, input)
}

func (template *SeqTemplate) GetSqlState(tplOpts *xdb.TemplateOptions) xdb.SqlState {
	sqlState := template.sqlStatePool.Get().(xdb.SqlState)
	sqlState.WithPlaceholder(template.Placeholder())
	sqlState.WithTemplateOptions(tplOpts)
	return sqlState
}

func (template *SeqTemplate) ReleaseSqlState(state xdb.SqlState) {
	state.Reset()
	template.sqlStatePool.Put(state)
}
