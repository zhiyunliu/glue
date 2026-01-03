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
	NewSqlState func(Placeholder) SqlState
)

type ExprName interface {
	GetPropName() string
	GetOper() string
	GetMatcher() ExpressionMatcher
}

// SqlState 用户记录sql状态
type SqlState interface {
	GetNames() []ExprName
	GetValues() []any
	UseExprCache() bool
	SetDynamic(DynamicType)
	HasDynamic(DynamicType) bool
	AppendExpr(exprName ExprName, value any) (phName string)
	CanCache() bool
	BuildCache(sql string) ExpressionCache
	WithPlaceholder(Placeholder)
	WithTemplateOptions(*TemplateOptions)
	Reset()
}

type SqlStatePool interface {
	Get() SqlState
	Put(state SqlState)
}
