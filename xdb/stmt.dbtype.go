package xdb

type StmtDbTypeHandler interface {
	Name() string
	//args:a=b => [a,b]
	Handle(param any, args []string) any
}

type StmtDbTypeProcessor interface {
	// RegistHandler 注册表达式匹配器
	RegistHandler(handler ...StmtDbTypeHandler)
	Process(param any, tagOpts TagOptions) any
}

var (
	// 新建一个请求参数处理器
	NewStmtDbTypeProcessor func(matchers ...StmtDbTypeHandler) StmtDbTypeProcessor
)
