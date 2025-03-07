package expression

import "github.com/zhiyunliu/glue/xdb"

type replaceSymbols struct{}

func (s *replaceSymbols) Name() string {
	return xdb.SymbolReplace
}

func (s *replaceSymbols) DynamicType() xdb.DynamicType {
	return xdb.DynamicReplace
}

func (s *replaceSymbols) Concat() string {
	return ""
}
func (s *replaceSymbols) IsDynamic() bool {
	return true
}

// func (s *replaceSymbols) Callback(item xdb.SqlState, valuer xdb.ExpressionValuer, input xdb.DBParam) (string, xdb.MissError) {
// 	item.SetDynamic(s.DynamicType())
// 	propName := valuer.GetPropName()
// 	argName, value, _ := input.Get(propName, item.GetPlaceholder())

// 	if !xdb.IsNil(value) {
// 		return valuer.Build(item, input, argName, value)
// 	}
// 	return "", nil
// }
