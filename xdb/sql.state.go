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

type SqlState interface {
	GetPlaceholder() Placeholder
	GetNames() []string
	GetValues() []any
	UseExprCache() bool
	SetDynamic(DynamicType)
	HasDynamic(DynamicType) bool
	BuildCache() SqlStateCahe
	AppendExpr(propName string, value any)
}

type DefaultSqlState struct {
	tplOpts     *TemplateOptions
	Names       []string
	Values      []any
	Placeholder Placeholder
	DynamicType DynamicType
}

func NewDefaultSqlState(ph Placeholder, tplOpts *TemplateOptions) SqlState {
	return &DefaultSqlState{
		tplOpts:     tplOpts,
		Placeholder: ph,
	}
}

func (s *DefaultSqlState) GetPlaceholder() Placeholder {
	return s.Placeholder
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
		ph: s.Placeholder,
	}
}

func (s *DefaultSqlState) AppendExpr(propName string, value any) {
	s.Names = append(s.Names, propName)
	s.Values = append(s.Values, value)
}
