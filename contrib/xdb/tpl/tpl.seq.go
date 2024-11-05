package tpl

import (
	"fmt"

	"github.com/zhiyunliu/glue/xdb"
)

// SeqTemplate 参数化时使用@+参数名作为占位符的SQL数据库如:oracle,sql server
type SeqTemplate struct {
	name    string
	prefix  string
	symbols xdb.SymbolMap
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

func (ph *seqPlaceHolder) Clone() xdb.Placeholder {
	return &seqPlaceHolder{
		idx:      ph.idx,
		template: ph.template,
	}
}

func NewSeq(name, prefix string, matcher TemplateMatcher) xdb.SQLTemplate {
	return &SeqTemplate{
		name:    name,
		prefix:  prefix,
		symbols: defaultSymbols.Clone(),
	}
}

func (template SeqTemplate) Name() string {
	return template.name
}

// GetSQLContext 获取查询串
func (template *SeqTemplate) GetSQLContext(tpl string, input map[string]interface{}) (sql string, args []any, err error) {
	return AnalyzeTPLFromCache(template, tpl, input, template.Placeholder())
}

func (template *SeqTemplate) Placeholder() xdb.Placeholder {
	return &seqPlaceHolder{template: template, idx: 0}
}

func (template *SeqTemplate) RegistPropertyMatcher(matcher ...xdb.ExpressionMatcher) error {
	return nil
}

func (template *SeqTemplate) RegistSymbol(symbols ...xdb.Symbol) error {
	return nil
}
