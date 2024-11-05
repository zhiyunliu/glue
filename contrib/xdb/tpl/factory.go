package tpl

var (
	DefaultPatternList []string = []string{
		`([@|\|&|$]{\w+(\.\w+)?})`,
		`([&|\|](({(like\s+%?\w+(\.\w+)*%?}))|({\w+(\.\w+)?\s+like\s+%?\w+%?})))`,
		`([&|\|](({in\s+\w+(\.\w+)*(=\w+)?\})|({\w+(\.\w+)?\s+in\s+\w+})))`,
		`([&|\|](({(>|>=|<|<=)\s+\w+(\.\w+)?})|({(\w+\.)?\w+(>|>=|<|<=)\w+})))`,
	}
)

const (
	//TotalPattern = `(@\{\w+(\.\w+)?\})|([&|\|]\{like\s+%?\w+(\.\w+)*%?})|([&|\|]\{in\s+\w+(\.\w+)*(=\w+)?\})|([&|\|]\{((>|>=|<|<=)\s+)?\w+(\.\w+)*(=\w+)?\})`
	// ParamPattern   = `[@]\{\w*[\.]?\w+\}`
	// AndPattern     = `[&]\{\w*[\.]?\w+\}`
	// OrPattern      = `[\|]\{\w*[\.]?\w+\}`

	//替换
	ReplacePattern = `\$\{\w*[\.]?\w+\}`

	SymbolAt      = "@"
	SymbolAnd     = "&"
	SymbolOr      = "|"
	SymbolReplace = "$"
)
