package expression

import "github.com/zhiyunliu/glue/xdb"

func initSqlState() {
	xdb.NewSqlState = NewDefaultSqlState
}

type DefaultSqlState struct {
	tplOpts     *xdb.TemplateOptions
	names       []xdb.ExprName
	values      []any
	placeholder xdb.Placeholder
	dynamicType xdb.DynamicType
}

func NewDefaultSqlState(ph xdb.Placeholder) xdb.SqlState {
	return &DefaultSqlState{
		placeholder: ph,
	}
}

func (s *DefaultSqlState) GetNames() []xdb.ExprName {
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

func (s *DefaultSqlState) AppendExpr(exprValuer xdb.ExprName, value any) (phName string) {

	propName := exprValuer.GetPropName()

	argName, phName := s.placeholder.Get(propName)
	value = s.placeholder.BuildArgVal(argName, value)

	s.names = append(s.names, exprValuer)
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

func (s *DefaultSqlState) WithArgs(placeholder xdb.Placeholder, tplOpts *xdb.TemplateOptions) {
	s.placeholder = placeholder
	s.tplOpts = tplOpts
}

func (s *DefaultSqlState) WithPlaceholder(placeholder xdb.Placeholder) {
	s.placeholder = placeholder
}

func (s *DefaultSqlState) WithTemplateOptions(tplOpts *xdb.TemplateOptions) {
	s.tplOpts = tplOpts
}

func (s *DefaultSqlState) Reset() {
	s.tplOpts = nil
	s.names = nil
	s.values = nil
	s.dynamicType = xdb.DynamicNone
}

type defaultSqlTemplateCache struct {
	sql   string
	names []xdb.ExprName
}

func (stc *defaultSqlTemplateCache) Build(state xdb.SqlState, input xdb.DBParam) (sql string, err error) {
	for _, expr := range stc.names {
		val, err := input.GetVal(expr.GetPropName())
		if err != nil {
			return "", err
		}

		operator, ok := expr.GetMatcher().GetOperatorMap().Load(expr.GetOper())
		if !ok {
			state.AppendExpr(expr, val)

		} else {
			newVal, err := operator.NormalizeValue(expr, input, val)
			if err != nil {
				return "", err
			}
			state.AppendExpr(expr, newVal)
		}
	}
	return stc.sql, nil
}
