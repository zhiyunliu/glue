package expression

import (
	"github.com/zhiyunliu/glue/xdb"
)

type atSymbols struct{}

func (s *atSymbols) Name() string {
	return xdb.SymbolAt
}

func (s *atSymbols) Concat() string {
	return ""
}
func (s *atSymbols) Callback(item xdb.SqlState, valuer xdb.ExpressionValuer, input xdb.DBParam) (string, xdb.MissError) {

	propName := valuer.GetPropName()

	argName, value, err := input.Get(propName, item.GetPlaceholder())
	if err != nil {
		return "", err
	}

	if !xdb.IsNil(value) {
		item.AppendExpr(propName, value)
	} else {
		item.AppendExpr(propName, nil)
	}
	return argName, nil

}
