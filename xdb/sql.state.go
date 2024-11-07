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
	CanCache() bool
	SetDynamic(DynamicType)
	HasDynamic(DynamicType) bool
	BuildCache() SqlStateCahe
	AppendExpr(name string, value any)
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
func (s *DefaultSqlState) CanCache() bool {
	return s.DynamicType&DynamicAnd > 0 ||
		s.DynamicType&DynamicOr > 0 ||
		s.DynamicType&DynamicReplace > 0
}
func (s *DefaultSqlState) SetDynamic(val DynamicType) {
	s.DynamicType = s.DynamicType | val
}
func (s *DefaultSqlState) HasDynamic(val DynamicType) bool {
	return s.DynamicType&val > 0
}
func (s *DefaultSqlState) BuildCache() SqlStateCahe {
	return &DefaultSqlStateCahe{}
}
func (s *DefaultSqlState) AppendExpr(name string, value any) {
	s.Names = append(s.Names, name)
	s.Values = append(s.Values, value)
}
