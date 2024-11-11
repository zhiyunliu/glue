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
	CanCache() bool
	BuildCache(sql string) SQLTemplateCache
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
func (s *DefaultSqlState) CanCache() bool {
	return !(s.HasDynamic(DynamicAnd) ||
		s.HasDynamic(DynamicOr) ||
		s.HasDynamic(DynamicReplace))
}

func (s *DefaultSqlState) BuildCache(sql string) SQLTemplateCache {
	return &defaultSqlTemplateCache{
		sql:   sql,
		names: s.names,
	}
}

type defaultSqlTemplateCache struct {
	sql   string
	names []string
}

func (stc *defaultSqlTemplateCache) Build(state SqlState, input DBParam) (sql string, err error) {
	for _, name := range stc.names {
		val, err := input.GetVal(name)
		if err != nil {
			return "", err
		}
		state.AppendExpr(name, val)
	}
	return stc.sql, nil
}
