package expression

import (
	"fmt"
	"strings"

	"github.com/zhiyunliu/glue/xdb"
)

func init() {
	var normalOperMap = NewOperatorMap()

	normalOperMap.Store("$", func(param xdb.DBParam, item xdb.ExpressionValuer, concat, argName string) string {
		value, err := param.GetVal(item.GetPropName())
		if err != nil {
			return ""
		}
		switch t := value.(type) {
		case []int8, []int, []int16, []int32, []int64, []uint, []uint16, []uint32, []uint64:
			value = strings.Trim(strings.Replace(fmt.Sprint(t), " ", ",", -1), "[]")
		case []string:
			if len(t) == 0 {
				return ""
			}
			value = "'" + strings.Join(t, "','") + "'"
		}
		if xdb.IsNil(value) {
			return ""
		}
		return fmt.Sprintf("%v", value)
	})

}
