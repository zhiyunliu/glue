package expression

import "github.com/zhiyunliu/glue/xdb"

type orSymbols struct{}

func (s *orSymbols) Name() string {
	return xdb.SymbolOr
}

func (s *orSymbols) DynamicType() xdb.DynamicType {
	return xdb.DynamicOr
}

func (s *orSymbols) Concat() string {
	return "or"
}
func (s *orSymbols) IsDynamic() bool {
	return true
}

// func (s *orSymbols) Callback(item xdb.SqlState, valuer xdb.ExpressionValuer, input xdb.DBParam) (string, xdb.MissError) {
// 	item.SetDynamic(s.DynamicType())

// 	propName := valuer.GetPropName()
// 	argName, value, _ := input.Get(propName, item.GetPlaceholder())

// 	if !xdb.IsNil(value) {
// 		return valuer.Build(item, input, argName, value)
// 	}
// 	return "", nil
// }
