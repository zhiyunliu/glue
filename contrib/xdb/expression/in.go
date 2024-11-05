package expression

import (
	"fmt"

	"github.com/zhiyunliu/glue/xdb"
)

func getInOperatorMap() OperatorMap {
	likeoperMap := NewOperatorMap()

	likeoperMap.Store("in", func(param xdb.DBParam, item xdb.ExpressionValuer, concat, argName string) string {
		return fmt.Sprintf("%s %s in (select value from string_split(%s,','))", concat, item.GetFullfield(), argName)
	})

	return likeoperMap
}
