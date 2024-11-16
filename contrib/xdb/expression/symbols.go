package expression

import (
	"github.com/zhiyunliu/glue/xdb"
)

// 根据表达式获取
var DefaultSymbols xdb.SymbolMap

func initSymbols() {
	DefaultSymbols = xdb.NewSymbolMap(
		&andSymbols{},
		&atSymbols{},
		&orSymbols{},
		&replaceSymbols{},
	)
}
