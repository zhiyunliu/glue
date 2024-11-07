package expression

import "github.com/zhiyunliu/glue/xdb"

type replaceSymbols struct{}

func (s *replaceSymbols) Name() string {
	return xdb.SymbolAnd
}

func (s *replaceSymbols) Concat() string {
	return ""
}
func (s *replaceSymbols) Callback(item xdb.SqlState, valuer xdb.ExpressionValuer, input xdb.DBParam) (string, xdb.MissError) {
	item.SetDynamic(xdb.DynamicReplace)

	propName := valuer.GetPropName()

	argName, value, _ := input.Get(propName, item.GetPlaceholder())

	if !xdb.IsNil(value) {
		item.AppendExpr(propName, value)
		return valuer.Build(input, argName)
	}
	return "", nil
}
