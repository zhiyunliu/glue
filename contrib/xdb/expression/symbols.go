package expression

import (
	"github.com/zhiyunliu/glue/xdb"
)

// 根据表达式获取
var GetExpressionValuer func(expression string, opts *xdb.ExpressionOptions) (matcher xdb.ExpressionValuer)

var DefaultSymbols xdb.SymbolMap = xdb.NewSymbolMap(
	&andSymbols{},
	&atSymbols{},
	&orSymbols{},
	&replaceSymbols{},
)
