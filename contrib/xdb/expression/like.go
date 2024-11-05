package expression

import (
	"fmt"

	"github.com/zhiyunliu/glue/xdb"
)

func getLikeOperatorMap() OperatorMap {
	likeoperMap := NewOperatorMap()

	likeoperMap.Store("like", func(param xdb.DBParam, item xdb.ExpressionValuer, concat, argName string) string {
		return fmt.Sprintf("%s %s like %s", concat, item.GetFullfield(), argName)
	})

	likeoperMap.Store("%like", func(param xdb.DBParam, item xdb.ExpressionValuer, concat, argName string) string {
		return fmt.Sprintf("%s %s like '%%'+%s", concat, item.GetFullfield(), argName)
	})

	likeoperMap.Store("like%", func(param xdb.DBParam, item xdb.ExpressionValuer, concat, argName string) string {
		return fmt.Sprintf("%s %s like %s+'%%'", concat, item.GetFullfield(), argName)
	})

	likeoperMap.Store("%like%", func(param xdb.DBParam, item xdb.ExpressionValuer, concat, argName string) string {
		return fmt.Sprintf("%s %s like '%%'+%s+'%%'", concat, item.GetFullfield(), argName)
	})
	return likeoperMap
}
