package sqlserver

import "github.com/zhiyunliu/glue/xdb"

type MssqlSqlState struct {
	innerState xdb.SqlState
	propCache  map[string]string
}

func NewSqlState(placeHolder xdb.Placeholder, tplOpts *xdb.TemplateOptions) xdb.SqlState {
	return &MssqlSqlState{
		innerState: xdb.NewDefaultSqlState(placeHolder, tplOpts),
		propCache:  make(map[string]string),
	}
}

func (s *MssqlSqlState) GetNames() []string {
	return s.innerState.GetNames()
}
func (s *MssqlSqlState) GetValues() []any {
	return s.innerState.GetValues()
}
func (s *MssqlSqlState) UseExprCache() bool {
	return s.innerState.UseExprCache()
}
func (s *MssqlSqlState) SetDynamic(dynamicType xdb.DynamicType) {
	s.innerState.SetDynamic(dynamicType)
}
func (s *MssqlSqlState) HasDynamic(dynamicType xdb.DynamicType) bool {
	return s.innerState.HasDynamic(dynamicType)
}

func (s *MssqlSqlState) AppendExpr(propName string, value any) (phName string) {
	phName, ok := s.propCache[propName]
	if ok {
		return phName
	}
	phName = s.innerState.AppendExpr(propName, value)
	s.propCache[propName] = phName
	return
}
