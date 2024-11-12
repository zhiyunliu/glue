package xdb

type DynamicType int

const (
	DynamicNone    DynamicType = 0
	DynamicAnd     DynamicType = 1
	DynamicOr      DynamicType = 2
	DynamicReplace DynamicType = 4
)

var (
	//新建一个SqlState
	NewSqlState func(Placeholder, *TemplateOptions) SqlState
)

// SqlState 用户记录sql状态
type SqlState interface {
	GetNames() []string
	GetValues() []any
	UseExprCache() bool
	SetDynamic(DynamicType)
	HasDynamic(DynamicType) bool
	AppendExpr(propName string, value any) (phName string)
	CanCache() bool
	BuildCache(sql string) ExpressionCache
}
