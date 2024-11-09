package xdb

type DynamicType int

const (
	DynamicNone    DynamicType = 0
	DynamicAnd     DynamicType = 1
	DynamicOr      DynamicType = 2
	DynamicReplace DynamicType = 4
)

type SqlStateCahe interface {
	Build(DBParam) (execSql string, values []interface{}, err error)
}

// SqlState 用户记录sql状态
type SqlState interface {
	GetNames() []string
	GetValues() []any
	UseExprCache() bool
	SetDynamic(DynamicType)
	HasDynamic(DynamicType) bool
	//BuildCache() SqlStateCahe
	AppendExpr(propName string, value any) (phName string)
}

type DefaultSqlState struct {
	tplOpts     *TemplateOptions
	Names       []string
	Values      []any
	placeholder Placeholder
	DynamicType DynamicType
}

func NewDefaultSqlState(ph Placeholder, tplOpts *TemplateOptions) SqlState {
	return &DefaultSqlState{
		tplOpts:     tplOpts,
		placeholder: ph,
	}
}

func (s *DefaultSqlState) GetNames() []string {
	return s.Names
}

func (s *DefaultSqlState) GetValues() []any {
	return s.Values
}

func (s *DefaultSqlState) UseExprCache() bool {
	return s.tplOpts.UseExprCache
}

func (s *DefaultSqlState) SetDynamic(val DynamicType) {
	s.DynamicType = s.DynamicType | val
}

func (s *DefaultSqlState) HasDynamic(val DynamicType) bool {
	return s.DynamicType&val > 0
}

func (s *DefaultSqlState) BuildCache() SqlStateCahe {
	return &DefaultSqlStateCahe{
		ph: s.placeholder,
	}
}

func (s *DefaultSqlState) AppendExpr(propName string, value any) (phName string) {
	argName, phName := s.placeholder.Get(propName)
	value = s.placeholder.BuildArgVal(argName, value)

	s.Names = append(s.Names, propName)
	s.Values = append(s.Values, value)
	return phName
}
