package expression

import "github.com/zhiyunliu/glue/xdb"

func initSqlState() {
	xdb.NewSqlState = NewDefaultSqlState
}

type DefaultSqlState struct {
	tplOpts     *xdb.TemplateOptions
	names       []string
	values      []any
	placeholder xdb.Placeholder
	dynamicType xdb.DynamicType
}

func NewDefaultSqlState(ph xdb.Placeholder, tplOpts *xdb.TemplateOptions) xdb.SqlState {
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

func (s *DefaultSqlState) SetDynamic(val xdb.DynamicType) {
	s.dynamicType = s.dynamicType | val
}

func (s *DefaultSqlState) HasDynamic(val xdb.DynamicType) bool {
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
	return !(s.HasDynamic(xdb.DynamicAnd) ||
		s.HasDynamic(xdb.DynamicOr) ||
		s.HasDynamic(xdb.DynamicReplace))
}

func (s *DefaultSqlState) BuildCache(sql string) xdb.ExpressionCache {
	return &defaultSqlTemplateCache{
		sql:   sql,
		names: s.names,
	}
}

type defaultSqlTemplateCache struct {
	sql   string
	names []string
}

func (stc *defaultSqlTemplateCache) Build(state xdb.SqlState, input xdb.DBParam) (sql string, err error) {
	for _, name := range stc.names {
		val, err := input.GetVal(name)
		if err != nil {
			return "", err
		}
		state.AppendExpr(name, val)
	}
	return stc.sql, nil
}
