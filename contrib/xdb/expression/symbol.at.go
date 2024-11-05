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
func (s *atSymbols) Callback(item *xdb.SqlScene, valuer xdb.ExpressionValuer, input xdb.DBParam) (string, xdb.MissError) {

	propName := valuer.GetPropName()

	argName, value, err := input.Get(propName, item.Placeholder)
	if err != nil {
		return "", err
	}
	if !xdb.IsNil(value) {
		item.Names = append(item.Names, propName)
		item.Values = append(item.Values, value)
	} else {
		item.Names = append(item.Names, propName)
		item.Values = append(item.Values, nil)
	}
	return argName, nil

}
