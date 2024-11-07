package expression

import "github.com/zhiyunliu/glue/xdb"

type andSymbols struct{}

func (s *andSymbols) Name() string {
	return xdb.SymbolAnd
}

func (s *andSymbols) DynamicType() xdb.DynamicType {
	return xdb.DynamicAnd
}

func (s *andSymbols) Concat() string {
	return "and"
}

func (s *andSymbols) Callback(item xdb.SqlState, valuer xdb.ExpressionValuer, input xdb.DBParam) (string, xdb.MissError) {
	item.SetDynamic(s.DynamicType())

	propName := valuer.GetPropName()
	argName, value, _ := input.Get(propName, item.GetPlaceholder())
	if !xdb.IsNil(value) {
		return valuer.Build(item, input, argName, value)
	}
	return "", nil
}
