package xdb

type DynamicType int

const (
	DynamicNone    DynamicType = 0
	DynamicAnd     DynamicType = 1
	DynamicOr      DynamicType = 2
	DynamicReplace DynamicType = 4
)

// SqlState 用户记录sql状态
type SqlState interface {
	GetNames() []string
	GetValues() []any
	UseExprCache() bool
	SetDynamic(DynamicType)
	HasDynamic(DynamicType) bool
	AppendExpr(propName string, value any) (phName string)
}

type DefaultSqlState struct {
	tplOpts     *TemplateOptions
	names       []string
	values      []any
	placeholder Placeholder
	dynamicType DynamicType
}

func NewDefaultSqlState(ph Placeholder, tplOpts *TemplateOptions) SqlState {
	return &DefaultSqlState{
		tplOpts:     tplOpts,
		placeholder: ph,
	}
}

func (s *DefaultSqlState) GetNames() []string {
	return s.names
}

func (s *DefaultSqlState) GetValues() []any {
	return s.values
}

func (s *DefaultSqlState) UseExprCache() bool {
	return s.tplOpts.UseExprCache
}

func (s *DefaultSqlState) SetDynamic(val DynamicType) {
	s.dynamicType = s.dynamicType | val
}

func (s *DefaultSqlState) HasDynamic(val DynamicType) bool {
	return s.dynamicType&val > 0
}

func (s *DefaultSqlState) AppendExpr(propName string, value any) (phName string) {
	argName, phName := s.placeholder.Get(propName)
	value = s.placeholder.BuildArgVal(argName, value)

	s.names = append(s.names, propName)
	s.values = append(s.values, value)
	return phName
}
