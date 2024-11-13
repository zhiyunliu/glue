package sqlserver

import (
	"database/sql"

	"github.com/zhiyunliu/glue/xdb"
)

type MssqlSqlState struct {
	innerState  xdb.SqlState
	placeHolder xdb.Placeholder
	phNameCache map[string]string
}

// 新建一个SqlState
func NewSqlState(placeHolder xdb.Placeholder) xdb.SqlState {
	return &MssqlSqlState{
		innerState:  xdb.NewSqlState(placeHolder),
		placeHolder: placeHolder,
		phNameCache: make(map[string]string),
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
	phName, ok := s.phNameCache[propName]
	if ok {
		return phName
	}

	var argPhName string
	if value != nil {
		if tmpv, ok := value.(sql.NamedArg); ok {
			value = tmpv
			argPhName = s.placeHolder.NamedArg(tmpv.Name)
		}
	}
	phName = s.innerState.AppendExpr(propName, value)
	if argPhName != "" {
		phName = argPhName
	}
	s.phNameCache[propName] = phName
	return
}

func (s *MssqlSqlState) CanCache() bool {
	return s.innerState.CanCache()
}

func (s *MssqlSqlState) BuildCache(sql string) xdb.ExpressionCache {
	return s.innerState.BuildCache(sql)
}

func (s *MssqlSqlState) WithPlaceholder(placeholder xdb.Placeholder) {
	s.innerState.WithPlaceholder(placeholder)
}

func (s *MssqlSqlState) WithTemplateOptions(tplOpts *xdb.TemplateOptions) {
	s.innerState.WithTemplateOptions(tplOpts)
}
func (s *MssqlSqlState) Reset() {
	clear(s.phNameCache)
	s.innerState.Reset()
}
